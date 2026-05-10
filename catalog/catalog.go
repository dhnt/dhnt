// Copyright 2026 The dhnt Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0

package catalog

import (
	"embed"
	"fmt"
	"io/fs"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"gopkg.in/yaml.v3"
)

// Skill is one entry in the catalog. Built once at package init from
// an embedded markdown file; treat as immutable.
type Skill struct {
	Name          string   // unique slug; matches the directory name
	Description   string   // one-line summary
	Phase         string   // SDLC phase tag; one of validPhases
	Executor      string   // markdown | builtin | cnl
	UserInvocable bool     // exposed via /<name>
	Aliases       []string // optional synonyms (case-insensitive lookup)
	Origin        string   // provenance: priorart/<tool> or "new"
	Body          string   // markdown body (frontmatter stripped)
	Path          string   // logical path within the catalog tree, e.g. "review/code-review"
}

// Phase tags accepted in frontmatter. The ordering reflects the
// chronological lifecycle from greenfield discovery through to
// production maintenance.
var validPhases = []string{
	"discover",
	"plan",
	"build",
	"test",
	"review",
	"commit",
	"integrate",
	"release",
	"deploy",
	"operate",
	"maintain",
	"document",
	"onboard",
}

// Executors recognised in frontmatter.
var validExecutors = map[string]bool{
	"markdown": true,
	"builtin":  true,
	"cnl":      true,
}

//go:embed all:md
var catalogFS embed.FS

var (
	loadOnce  sync.Once
	loadErr   error
	all       []Skill
	byName    map[string]*Skill // includes aliases
	phaseList []string
)

// All returns every skill in the catalog, sorted by Phase then Name.
// The returned slice is owned by the package; callers must not
// mutate it.
func All() []Skill {
	mustLoad()
	return all
}

// Count returns the number of skills in the catalog.
func Count() int {
	mustLoad()
	return len(all)
}

// Lookup returns the skill registered under the given name (or any
// of its aliases). Lookup is case-insensitive after whitespace
// trimming.
func Lookup(name string) (Skill, bool) {
	mustLoad()
	s, ok := byName[normalise(name)]
	if !ok {
		return Skill{}, false
	}
	return *s, true
}

// ByPhase returns all skills in the given phase, sorted by Name.
// Returns an empty slice for unknown phases.
func ByPhase(phase string) []Skill {
	mustLoad()
	var out []Skill
	for _, s := range all {
		if s.Phase == phase {
			out = append(out, s)
		}
	}
	return out
}

// Phases returns the set of phase tags that have at least one skill,
// in canonical lifecycle order.
func Phases() []string {
	mustLoad()
	return phaseList
}

// LoadError exposes any error encountered during the lazy load. Most
// callers don't need this — the package-level functions panic on
// load failure since the embedded catalog is a build-time artifact.
// Tests use this to assert structural invariants without the panic.
func LoadError() error {
	mustLoad()
	return loadErr
}

func mustLoad() {
	loadOnce.Do(load)
	if loadErr != nil {
		// The embedded catalog should have been validated by
		// catalog_test.go before any tag was cut. If we still hit
		// an error at runtime, it is a programmer bug, not a user
		// bug.
		panic("dhnt/catalog: load failed: " + loadErr.Error())
	}
}

func load() {
	all = nil
	byName = make(map[string]*Skill)
	seenPhases := make(map[string]bool)

	walkErr := fs.WalkDir(catalogFS, "md", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if filepath.Base(path) != "skill.md" {
			return nil
		}
		data, err := catalogFS.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read %s: %w", path, err)
		}
		s, err := parseSkill(path, data)
		if err != nil {
			return fmt.Errorf("parse %s: %w", path, err)
		}
		all = append(all, s)
		seenPhases[s.Phase] = true
		return nil
	})
	if walkErr != nil {
		loadErr = walkErr
		return
	}

	// Stable sort: phase first (canonical lifecycle order), then name.
	phaseRank := make(map[string]int, len(validPhases))
	for i, p := range validPhases {
		phaseRank[p] = i
	}
	sort.SliceStable(all, func(i, j int) bool {
		ri, rj := phaseRank[all[i].Phase], phaseRank[all[j].Phase]
		if ri != rj {
			return ri < rj
		}
		return all[i].Name < all[j].Name
	})

	// Build name index after sorting so the &Skill pointers are stable.
	for i := range all {
		s := &all[i]
		if existing, dup := byName[normalise(s.Name)]; dup {
			loadErr = fmt.Errorf("duplicate skill name %q (also at %s)", s.Name, existing.Path)
			return
		}
		byName[normalise(s.Name)] = s
		for _, a := range s.Aliases {
			if existing, dup := byName[normalise(a)]; dup {
				loadErr = fmt.Errorf("alias %q for skill %q collides with %s",
					a, s.Name, existing.Path)
				return
			}
			byName[normalise(a)] = s
		}
	}

	// Phases in canonical order, only those that have entries.
	for _, p := range validPhases {
		if seenPhases[p] {
			phaseList = append(phaseList, p)
		}
	}
}

func parseSkill(path string, data []byte) (Skill, error) {
	body := string(data)
	const marker = "---\n"
	if !strings.HasPrefix(body, marker) {
		return Skill{}, fmt.Errorf("missing leading frontmatter marker")
	}
	rest := body[len(marker):]
	end := strings.Index(rest, "\n"+marker[:len(marker)-1]+"\n")
	if end < 0 {
		// allow either form: --- followed by newline or EOF
		end = strings.Index(rest, "\n---\n")
	}
	if end < 0 {
		return Skill{}, fmt.Errorf("missing closing frontmatter marker")
	}
	frontmatter := rest[:end]
	mdBody := strings.TrimLeft(rest[end+len("\n---\n"):], "\n")

	var fm struct {
		Name          string   `yaml:"name"`
		Description   string   `yaml:"description"`
		Phase         string   `yaml:"phase"`
		Executor      string   `yaml:"executor"`
		UserInvocable *bool    `yaml:"user_invocable"`
		Aliases       []string `yaml:"aliases"`
		Origin        string   `yaml:"origin"`
	}
	if err := yaml.Unmarshal([]byte(frontmatter), &fm); err != nil {
		return Skill{}, fmt.Errorf("yaml: %w", err)
	}
	if fm.Name == "" {
		return Skill{}, fmt.Errorf("name is required")
	}
	if fm.Description == "" {
		return Skill{}, fmt.Errorf("description is required")
	}
	if !isValidPhase(fm.Phase) {
		return Skill{}, fmt.Errorf("phase %q is not one of %v", fm.Phase, validPhases)
	}
	if fm.Executor == "" {
		fm.Executor = "markdown"
	}
	if !validExecutors[fm.Executor] {
		return Skill{}, fmt.Errorf("executor %q is not one of markdown/builtin/cnl", fm.Executor)
	}
	if fm.Origin == "" {
		fm.Origin = "new"
	}

	// Verify the file path matches the declared name + phase.
	logicalPath := strings.TrimPrefix(path, "md/")
	logicalDir := filepath.Dir(logicalPath)
	parts := strings.Split(logicalDir, "/")
	if len(parts) != 2 {
		return Skill{}, fmt.Errorf("path must be md/<phase>/<name>/skill.md, got %s", path)
	}
	if parts[0] != fm.Phase {
		return Skill{}, fmt.Errorf("path phase %q does not match frontmatter phase %q", parts[0], fm.Phase)
	}
	if parts[1] != fm.Name {
		return Skill{}, fmt.Errorf("path name %q does not match frontmatter name %q", parts[1], fm.Name)
	}

	userInvocable := true
	if fm.UserInvocable != nil {
		userInvocable = *fm.UserInvocable
	}

	return Skill{
		Name:          fm.Name,
		Description:   fm.Description,
		Phase:         fm.Phase,
		Executor:      fm.Executor,
		UserInvocable: userInvocable,
		Aliases:       fm.Aliases,
		Origin:        fm.Origin,
		Body:          mdBody,
		Path:          logicalDir,
	}, nil
}

func isValidPhase(p string) bool {
	for _, v := range validPhases {
		if v == p {
			return true
		}
	}
	return false
}

func normalise(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}
