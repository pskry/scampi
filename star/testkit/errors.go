// SPDX-License-Identifier: GPL-3.0-only

package testkit

import "scampi.dev/scampi/testkit"

// The diagnostic types live in the top-level testkit package now.
// These aliases keep the legacy Starlark code path compiling until
// Phase 6 deletes star/ entirely.
type (
	TestPass    = testkit.TestPass
	TestFail    = testkit.TestFail
	TestSummary = testkit.TestSummary
	TestError   = testkit.TestError
)
