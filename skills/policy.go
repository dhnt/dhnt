// Copyright 2026 The dhnt Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0

package skills

// RunPolicy executes a skill honouring its declared OnFail policy. It is a
// thin driver over the pure single-shot Run; the executor itself never
// loops or swallows failure — the recovery intent lives in the skill and
// is interpreted here.
//
//   - PolicyAbort (default): run once; the result (valid or not) stands.
//   - PolicyRetry: re-run on an invalid result up to maxRetry extra times.
//     This only helps non-deterministic / eventual-consistency runs (e.g.
//     "wait for the fleet to converge"); identical deterministic steps will
//     just produce the same verdict. A hard runtime error stops the loop.
//   - PolicyBlockers: run once; an *invalid* result is returned WITHOUT an
//     error (graceful) so the caller can surface blockers and continue —
//     only a hard runtime error propagates.
//
// The returned RunOutcome is OutcomeBaseline when the run is valid and
// OutcomeFailed otherwise.
func RunPolicy(s Skill, env Env, tier string, maxRetry int) (Attestation, RunOutcome, error) {
	att, err := Run(s, env, tier)

	switch s.OnFail {
	case PolicyRetry:
		for i := 0; i < maxRetry && err == nil && !att.Valid; i++ {
			att, err = Run(s, env, tier)
		}
	case PolicyBlockers:
		// A failed contract is not an error under this policy; only a hard
		// runtime error (e.g. a missing binding) propagates.
		if err == nil && !att.Valid {
			return att, OutcomeFailed, nil
		}
	}

	if err != nil {
		return att, OutcomeFailed, err
	}
	if !att.Valid {
		return att, OutcomeFailed, nil
	}
	return att, OutcomeBaseline, nil
}
