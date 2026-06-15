// Copyright 2026 The dhnt Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0

package dev_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/dhnt/dhnt/skills"
	"github.com/dhnt/dhnt/skills/dev"
)

func git(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
	return string(out)
}

func initRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	git(t, dir, "init", "-q")
	git(t, dir, "config", "user.email", "d@e.com")
	git(t, dir, "config", "user.name", "d")
	// an initial commit so HEAD exists
	if err := os.WriteFile(filepath.Join(dir, "seed"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	git(t, dir, "add", "seed")
	git(t, dir, "commit", "-q", "-m", "seed")
	return dir
}

func TestSafeCommit_GreenCommits(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("uses git + POSIX true")
	}
	dir := initRepo(t)
	if err := os.WriteFile(filepath.Join(dir, "a.txt"), []byte("hello\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	// test command stubbed to `true` so the test is hermetic.
	spec := dev.SafeCommitSpec(dir, "feat: add a.txt", []string{"a.txt"}, "true")
	env, _, err := dev.NewEnv(spec)
	if err != nil {
		t.Fatal(err)
	}
	att, err := skills.Run(dev.SafeCommitSkill(), env, "safe-commit")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !att.Valid {
		t.Fatalf("expected valid commit; failed=%v", att.Failed)
	}
	if log := git(t, dir, "log", "-1", "--format=%s"); log != "feat: add a.txt\n" {
		t.Errorf("unexpected commit subject: %q", log)
	}
}

func TestSafeCommit_RedTestsDoNotCommit(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("uses git + POSIX false")
	}
	dir := initRepo(t)
	_ = os.WriteFile(filepath.Join(dir, "a.txt"), []byte("hello\n"), 0o644)
	before := git(t, dir, "rev-parse", "HEAD")
	spec := dev.SafeCommitSpec(dir, "should not land", []string{"a.txt"}, "false") // tests fail
	env, _, _ := dev.NewEnv(spec)
	if _, err := skills.Run(dev.SafeCommitSkill(), env, "safe-commit"); err == nil {
		t.Errorf("expected the run to fail before committing")
	}
	if after := git(t, dir, "rev-parse", "HEAD"); after != before {
		t.Errorf("a commit was created despite red tests")
	}
}

func TestPinBump_GuardsDirtySubmodule(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("uses git + sh")
	}
	// umbrella with a nested 'sub' git repo (stand-in for a submodule).
	umb := initRepo(t)
	sub := filepath.Join(umb, "sub")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	git(t, sub, "init", "-q")
	git(t, sub, "config", "user.email", "d@e.com")
	git(t, sub, "config", "user.name", "d")
	_ = os.WriteFile(filepath.Join(sub, "f"), []byte("v1"), 0o644)
	git(t, sub, "add", "f")
	git(t, sub, "commit", "-q", "-m", "v1")

	// dirty the submodule working tree → the guard must abort the bump.
	_ = os.WriteFile(filepath.Join(sub, "f"), []byte("dirty"), 0o644)
	before := git(t, umb, "rev-parse", "HEAD")
	spec := dev.PinBumpSpec(umb, "sub")
	env, _, _ := dev.NewEnv(spec)
	if _, err := skills.Run(dev.PinBumpSkill(), env, "pin-bump"); err == nil {
		t.Errorf("expected the guard to abort on a dirty submodule")
	}
	if after := git(t, umb, "rev-parse", "HEAD"); after != before {
		t.Errorf("a pin bump was committed despite a dirty submodule")
	}
}
