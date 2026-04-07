// SPDX-License-Identifier: GPL-3.0-only

package std

import (
	"testing"

	"scampi.dev/scampi/lang/check"
)

func TestStdLibCompiles(t *testing.T) {
	modules, err := check.BootstrapStd(FS)
	if err != nil {
		t.Fatalf("bootstrap: %v", err)
	}
	want := []string{"std", "posix", "rest", "container"}
	for _, name := range want {
		if _, ok := modules[name]; !ok {
			t.Errorf("missing module %q", name)
		}
	}
}
