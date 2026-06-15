// Copyright 2026 The dhnt Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0

// Command go_verify dogfoods the go-verify dhnt skill on a real Go module:
// it formats the tree, then verifies fmt-clean ∧ vet-clean ∧ tests-green,
// and prints the attestation. Point it at any module.
//
//	go run . [dir]                       # default: current directory
//	go run . ../some-module -- make test # override the test command
package main

import (
	"fmt"
	"os"

	"github.com/dhnt/dhnt/skills"
	"github.com/dhnt/dhnt/skills/dev"
)

func main() {
	dir := "."
	var testArgv []string
	args := os.Args[1:]
	if len(args) > 0 && args[0] != "--" {
		dir = args[0]
		args = args[1:]
	}
	if len(args) > 0 && args[0] == "--" {
		testArgv = args[1:] // everything after -- is the test command
	}

	spec := dev.GoVerifySpec(dir, testArgv...)
	env, sess, err := dev.NewEnv(spec)
	if err != nil {
		panic(err)
	}
	skill := dev.GoVerifySkill()
	id, _ := skills.Identity(skill)
	fmt.Printf("go-verify (%s) on %s\n", id[:14]+"…", dir)

	att, runErr := skills.Run(skill, env, "go-verify")
	if files := sess.Output("fi"); files != "" {
		fmt.Printf("unformatted:\n%s", files)
	}
	if runErr != nil {
		fmt.Printf("run error: %v\n", runErr)
		os.Exit(1)
	}
	fmt.Printf("valid=%v consistent=%v passed=%v failed=%v effects=%v\n",
		att.Valid, att.Consistent(skill), att.Passed, att.Failed, att.Effects)
	if !att.Valid {
		os.Exit(1)
	}
}
