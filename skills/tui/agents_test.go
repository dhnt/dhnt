// Copyright 2026 The dhnt Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0

package tui_test

import (
	"os/exec"
	"regexp"
	"runtime"
	"testing"

	"github.com/dhnt/dhnt/skills"
	"github.com/dhnt/dhnt/skills/tui"
)

// TestAgents_VersionSmoke drives each installed agent CLI's `--version`
// through the dhnt DriveOnce skill over a real PTY. It is token-free
// (no model call) yet exercises the full path against the real binaries:
// spawn → expect → clean-exit contract → attestation. Tools that are not
// installed are skipped.
func TestAgents_VersionSmoke(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("PTY driver is unix-only")
	}
	semver := regexp.MustCompile(`\d+\.\d+\.\d+`)
	for _, name := range []string{"claude", "codex", "gemini", "opencode", "aider"} {
		name := name
		t.Run(name, func(t *testing.T) {
			a := tui.Agents[name]
			if _, err := exec.LookPath(a.Version[0]); err != nil {
				t.Skipf("%s not on PATH", name)
			}
			spec, ok := tui.VersionSpec(name, "")
			if !ok {
				t.Fatalf("no spec for %q", name)
			}
			env, sess, err := tui.NewEnv(spec)
			if err != nil {
				t.Fatalf("NewEnv: %v", err)
			}
			defer sess.Close()

			att, err := skills.Run(tui.DriveOnceSkill(), env, name+"-version")
			if err != nil {
				t.Fatalf("Run: %v\noutput:\n%s", err, sess.Output())
			}
			if !att.Valid || !att.Consistent(tui.DriveOnceSkill()) {
				t.Errorf("%s version run not valid/consistent: %+v\noutput:\n%s", name, att, sess.Output())
			}
			if !semver.MatchString(sess.Output()) {
				t.Errorf("%s: no semver in output:\n%s", name, sess.Output())
			}
		})
	}
}
