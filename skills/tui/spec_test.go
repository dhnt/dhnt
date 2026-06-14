// Copyright 2026 The dhnt Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0

package tui_test

import (
	"testing"
	"time"

	"github.com/dhnt/dhnt/skills/tui"
)

func TestSpecFromJSON(t *testing.T) {
	data := []byte(`{"argv":["sh","-c","echo hi"],"dir":"/tmp","patterns":{"readayu":"hi"},"inputs":{"x":"y\n"},"quit":"exit\n","timeout_seconds":2.5}`)
	spec, err := tui.SpecFromJSON(data)
	if err != nil {
		t.Fatalf("SpecFromJSON: %v", err)
	}
	if len(spec.Argv) != 3 || spec.Argv[2] != "echo hi" {
		t.Errorf("argv: %v", spec.Argv)
	}
	if spec.Dir != "/tmp" || spec.Quit != "exit\n" {
		t.Errorf("dir/quit: %q %q", spec.Dir, spec.Quit)
	}
	if spec.Patterns["readayu"] != "hi" || spec.Inputs["x"] != "y\n" {
		t.Errorf("maps: %v %v", spec.Patterns, spec.Inputs)
	}
	if spec.Timeout != 2500*time.Millisecond {
		t.Errorf("timeout: %v", spec.Timeout)
	}
}
