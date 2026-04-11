// SPDX-License-Identifier: GPL-3.0-only

// Package testkit is the scampi-lang test framework runtime.
//
// It owns the bridge between scampi-lang test configs and Go-side
// mock targets:
//
//   - Matcher dispatch: each `matchers.*` constructor evaluates to a
//     [eval.StructVal]. Match interprets one of those against an
//     observed runtime value in a typed slot.
//   - Verifier: walks an `expect = test.ExpectedState{ ... }` value
//     produced by eval and runs Match on every slot/key, collecting
//     mismatches.
//   - (Future) Test runner: drives the lex → parse → check → eval →
//     resolve → apply → verify pipeline used by `scampi test`.
//
// See docs/test-framework.md for the design.
package testkit
