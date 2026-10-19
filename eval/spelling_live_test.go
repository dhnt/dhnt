// Copyright 2026 The dhnt Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0

package eval

import (
	"context"
	"flag"
	"os/exec"
	"strings"
	"testing"
	"time"
)

// Opt-in live smoke runner for the spelling eval (Sprint 476, story S1).
// Skipped by default; the manager invokes it for ONE smoke run on a
// subscription model. The prompt travels on the command's stdin and the
// model's completion is read from its stdout. Example:
//
//	go test ./eval/ -run TestSpellingLiveSmoke -spelling-live-cmd \
//	  "my-model-cli complete" -v
//
// Any wrapper that reads a prompt from stdin and prints a completion to
// stdout works (a shell pipeline, a subscription-model CLI, ...). This test
// never runs as part of the normal suite: without the flag it skips, and CI
// does not set the flag.
var spellingLiveCmd = flag.String("spelling-live-cmd", "",
	"shell command for ONE live smoke run: prompt on stdin, completion on stdout")

func TestSpellingLiveSmoke(t *testing.T) {
	if *spellingLiveCmd == "" {
		t.Skip("no -spelling-live-cmd given; live smoke run is opt-in")
	}
	model := func(prompt string) (string, error) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()
		cmd := exec.CommandContext(ctx, "sh", "-c", *spellingLiveCmd)
		cmd.Stdin = strings.NewReader(prompt)
		out, err := cmd.Output()
		return strings.TrimSpace(string(out)), err
	}
	verdicts, sum, err := RunSpelling(model, SpellingTasks())
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("live smoke results for %q:\n%s", *spellingLiveCmd,
		RenderSpellingMarkdown(verdicts, sum))
}
