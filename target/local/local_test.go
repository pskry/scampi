// SPDX-License-Identifier: GPL-3.0-only

package local

import (
	"context"
	"runtime"
	"testing"

	"scampi.dev/scampi/capability"
	"scampi.dev/scampi/spec"
)

func TestCreate_DetectsPkgBackend(t *testing.T) {
	tgt, err := Local{}.Create(context.Background(), nil, spec.TargetInstance{})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if !tgt.Capabilities().Has(capability.Pkg) {
		t.Fatalf(
			"expected local target on %s/%s to provide Pkg capability, but it didn't",
			runtime.GOOS,
			runtime.GOARCH,
		)
	}
}
