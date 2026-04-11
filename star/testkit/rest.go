// SPDX-License-Identifier: GPL-3.0-only

package testkit

import (
	"context"

	"scampi.dev/scampi/source"
	"scampi.dev/scampi/spec"
	"scampi.dev/scampi/target"
)

// RESTMockTargetType wraps a pre-built target.MemREST as a
// spec.TargetType so the Starlark eval pipeline can install it like
// any other target. The runtime mock itself lives in the target
// package as target.MemREST.
type RESTMockTargetType struct {
	Tgt *target.MemREST
}

func (t RESTMockTargetType) Kind() string   { return "rest_mock" }
func (t RESTMockTargetType) NewConfig() any { return nil }

func (t RESTMockTargetType) Create(
	_ context.Context,
	_ source.Source,
	_ spec.TargetInstance,
) (target.Target, error) {
	return t.Tgt, nil
}
