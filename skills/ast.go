// Copyright 2026 The dhnt Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0

package skills

// Skill is the Layer 2 AST root. All identifiers (Name, Caps,
// Contract predicates, Step names, primitive names, arg names, value
// references) are dhnt canonical strings keyed against the Glossary.
//
// The grammar covered today:
//
//	skill <name>
//	  needaso <cap>+ fini
//	  enisure <predicate> <arg-pairs> fini      // contract clause
//	  sotepo <step-name> <primitive> <arg-pairs> fini
//	fini
//
// Contract is the spine (pillar P1): a run is valid iff every Check
// holds (evaluates to bua). Steps are the optional implementation —
// any executor tier may reach the contracted end-state by a different
// path and is judged only by the contract. More complex shapes
// (effect caps, on-fail policies, nested flow control, intent refs)
// layer in by adding fields here without breaking the roundtrip.
type Skill struct {
	Name     string
	Caps     []string
	Contract []Check
	Steps    []Step
}

// Check is one contract clause (pillar P1): a predicate invoked with
// named arguments whose evaluation must yield true (bua) for a run to
// be valid. Predicate must be a dhnt-canonical glossary id of kind
// "predicate"; it is the typed, portable generalisation of an opaque
// verify command.
type Check struct {
	Predicate string
	Args      []Arg
}

// Step is one named action: it invokes a single primitive with named
// arguments. The name and primitive must be dhnt-canonical and exist
// in the bound Glossary.
type Step struct {
	Name      string
	Primitive string
	Args      []Arg
}

// Arg is a single name=value binding inside a primitive call. The
// name must be a dhnt-canonical glossary identifier; the value is an
// Expr.
type Arg struct {
	Name  string
	Value Expr
}

// Expr is a value expression. Phase 0 supports two variants:
//
//   - Ref:    a glossary reference (capability, type, named atom).
//   - Number: a non-negative decimal numeral.
//
// Free-text intent slots and richer expressions land in a follow-up.
type Expr struct {
	Kind   ExprKind
	Ref    string // Kind == ExprRef
	Number uint64 // Kind == ExprNumber
}

// ExprKind discriminates the Expr variants.
type ExprKind int

const (
	ExprInvalid ExprKind = iota
	ExprRef
	ExprNumber
)

// NewRef constructs a reference expression. The dhnt key is the
// caller's responsibility to validate against a Glossary.
func NewRef(dhnt string) Expr { return Expr{Kind: ExprRef, Ref: dhnt} }

// NewNumber constructs a numeric expression.
func NewNumber(n uint64) Expr { return Expr{Kind: ExprNumber, Number: n} }
