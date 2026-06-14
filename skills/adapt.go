// Copyright 2026 The dhnt Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0

package skills

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
)

// Runtime self-improvement (adaptation).
//
// A dhnt skill can learn from failure and reuse a verified fix — soundly,
// because the contract is a deterministic oracle. When a run fails in a
// given environment, a model proposes a corrected implementation; it is
// accepted ONLY if it still satisfies the *original* contract within the
// *original* effect cap; and it is then cached, keyed by (skill identity,
// context). A later call in the same context — even by a different agent —
// reuses the verified variant instead of repeating the failure. This is
// what a prose skill cannot do: it has no identity to key on and no oracle
// to certify a cached fix.

// EnvProbe is one declared environment fact that affects execution (OS,
// arch, a tool's version, a container digest). The set of probes defines
// what "the same context" means — declare the ones that matter, no more.
type EnvProbe struct {
	Name  string
	Value string
}

// ContextKey is a deterministic fingerprint of the environment.
func ContextKey(probes []EnvProbe) string {
	lines := make([]string, len(probes))
	for i, p := range probes {
		lines[i] = p.Name + "=" + p.Value
	}
	sort.Strings(lines)
	sum := sha256.Sum256([]byte(strings.Join(lines, "\n")))
	return "c" + hex.EncodeToString(sum[:])
}

// Adaptation is a verified, context-scoped variant of a skill. Variant is
// the canonical L1.5 of an implementation that satisfied the original
// skill's contract+cap; Attest is the proof.
type Adaptation struct {
	SkillID    string      `json:"skill_id"`
	ContextKey string      `json:"context_key"`
	Variant    string      `json:"variant"`
	Attest     Attestation `json:"attest"`
}

// AdaptationStore persists verified variants, shared across agents.
type AdaptationStore interface {
	Get(skillID, ctxKey string) (Adaptation, bool, error)
	Put(a Adaptation) error
}

// MemStore is an in-memory AdaptationStore (tests, ephemeral runs).
type MemStore struct {
	mu sync.Mutex
	m  map[string]Adaptation
}

func NewMemStore() *MemStore { return &MemStore{m: make(map[string]Adaptation)} }

func (s *MemStore) Get(skillID, ctxKey string) (Adaptation, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	a, ok := s.m[skillID+"/"+ctxKey]
	return a, ok, nil
}

func (s *MemStore) Put(a Adaptation) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.m[a.SkillID+"/"+a.ContextKey] = a
	return nil
}

// FileStore persists adaptations under <Dir>/<skillID>/<ctxKey>.json, so
// a fix learned by one agent is reused by any agent sharing the dir.
type FileStore struct{ Dir string }

func (s *FileStore) path(skillID, ctxKey string) string {
	return filepath.Join(s.Dir, skillID, ctxKey+".json")
}

func (s *FileStore) Get(skillID, ctxKey string) (Adaptation, bool, error) {
	data, err := os.ReadFile(s.path(skillID, ctxKey))
	if os.IsNotExist(err) {
		return Adaptation{}, false, nil
	}
	if err != nil {
		return Adaptation{}, false, err
	}
	var a Adaptation
	if err := json.Unmarshal(data, &a); err != nil {
		return Adaptation{}, false, err
	}
	return a, true, nil
}

func (s *FileStore) Put(a Adaptation) error {
	p := s.path(a.SkillID, a.ContextKey)
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(a, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(p, data, 0o644)
}

// Repairer proposes corrected implementations via a model. It adapts the
// *steps* only; the contract and effect cap are fixed by the skill.
type Repairer struct {
	Complete    Completer
	Glossary    *Glossary
	Lang        string
	MaxAttempts int
}

// RunOutcome classifies how an adaptive run resolved.
type RunOutcome int

const (
	OutcomeBaseline RunOutcome = iota // the skill ran as written
	OutcomeCached                     // a previously-learned variant ran
	OutcomeRepaired                   // a fresh fix was learned this run
	OutcomeFailed                     // could not satisfy the contract
)

func (o RunOutcome) String() string {
	switch o {
	case OutcomeBaseline:
		return "baseline"
	case OutcomeCached:
		return "cached"
	case OutcomeRepaired:
		return "repaired"
	default:
		return "failed"
	}
}

// RunAdaptive executes a skill with learning. It (1) reuses a cached
// verified variant for this (skill, context) if one re-verifies, else
// (2) runs the skill as written, else (3) asks the Repairer for a fix
// and accepts it ONLY if it satisfies the original contract within the
// original effect cap — then caches it. store and repair may be nil
// (degrading to a plain verified run).
func RunAdaptive(s Skill, env Env, tier string, probes []EnvProbe, store AdaptationStore, repair *Repairer) (Attestation, RunOutcome, error) {
	id, err := Identity(s)
	if err != nil {
		return Attestation{}, OutcomeFailed, err
	}
	ck := ContextKey(probes)

	// 1. cached variant — re-verify before trusting it (handles drift).
	if store != nil {
		if a, ok, _ := store.Get(id, ck); ok {
			if v, err := ParseDhnt(a.Variant); err == nil {
				if att, err := Run(v, env, tier); err == nil && att.Valid {
					return att, OutcomeCached, nil
				}
			}
		}
	}

	// 2. baseline.
	att, runErr := Run(s, env, tier)
	if runErr == nil && att.Valid {
		return att, OutcomeBaseline, nil
	}

	// 3. repair.
	if repair == nil || repair.Complete == nil {
		if runErr != nil {
			return Attestation{}, OutcomeFailed, runErr
		}
		return att, OutcomeFailed, fmt.Errorf("skills: contract not satisfied and no repairer")
	}
	maxA := repair.MaxAttempts
	if maxA <= 0 {
		maxA = 2
	}
	lastErr := runErr
	for attempt := 0; attempt < maxA; attempt++ {
		cand, perr := repair.propose(s, att, runErr)
		if perr != nil {
			lastErr = perr
			continue
		}
		// Graft the ORIGINAL contract + cap onto the candidate's steps:
		// the model may change the *how*, never the *what*. This is what
		// stops a model from "fixing" a run by weakening the spec.
		verify := Skill{
			Name:      s.Name,
			Caps:      s.Caps,
			EffectCap: s.EffectCap,
			Contract:  s.Contract,
			Steps:     cand.Steps,
		}
		vatt, verr := Run(verify, env, tier)
		if verr == nil && vatt.Valid {
			if store != nil {
				if canon, err := LineariseDhnt(verify); err == nil {
					_ = store.Put(Adaptation{SkillID: id, ContextKey: ck, Variant: canon, Attest: vatt})
				}
			}
			return vatt, OutcomeRepaired, nil
		}
		if verr != nil {
			lastErr = verr
		} else {
			lastErr = fmt.Errorf("candidate failed contract: %v / effects %v", vatt.Failed, vatt.Effects)
		}
	}
	return Attestation{}, OutcomeFailed, fmt.Errorf("skills: repair failed after %d attempts: %w", maxA, lastErr)
}

// propose asks the model for a corrected skill and parses it (loan-word
// mode, so free-form names are accepted).
func (r *Repairer) propose(s Skill, att Attestation, runErr error) (Skill, error) {
	lang := r.Lang
	if lang == "" {
		lang = "en"
	}
	failure := "the contract was not satisfied"
	if runErr != nil {
		failure = runErr.Error()
	} else if len(att.Failed) > 0 {
		failure = "failed checks: " + strings.Join(att.Failed, ", ")
	}
	out, err := r.Complete(r.repairPrompt(s, lang, failure))
	if err != nil {
		return Skill{}, err
	}
	cnl, ok := extractCNL(out)
	if !ok {
		return Skill{}, fmt.Errorf("repair: no <dhnt> block in model output")
	}
	return parseLangWith(cnl, r.Glossary, lang, true)
}

func (r *Repairer) repairPrompt(s Skill, lang, failure string) string {
	var b strings.Builder
	b.WriteString(NormalisePrompt(r.Glossary, lang))
	b.WriteString("\n\nA skill FAILED in the current environment. Produce a CORRECTED skill\n")
	b.WriteString("that achieves the SAME success criteria within the SAME effects but works\n")
	b.WriteString("here. Change only the steps (the how), never the success criteria.\n\n")
	if en, err := LineariseLang(s, r.Glossary, lang); err == nil {
		fmt.Fprintf(&b, "Failing skill: %s\n", en)
	}
	fmt.Fprintf(&b, "Failure: %s\n", failure)
	if len(s.Contract) > 0 {
		var crit []string
		for i := range s.Contract {
			crit = append(crit, s.Contract[i].Predicate)
		}
		fmt.Fprintf(&b, "Must still satisfy: %s\n", strings.Join(crit, ", "))
	}
	if len(s.EffectCap) > 0 {
		var effs []string
		for _, e := range s.EffectCap {
			effs = append(effs, e.String())
		}
		fmt.Fprintf(&b, "Allowed effects only: %s\n", strings.Join(effs, ", "))
	}
	return b.String()
}
