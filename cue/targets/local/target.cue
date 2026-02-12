// SPDX-License-Identifier: GPL-3.0-only

package local

#Target: {
	@doc("Local machine target - auto-detects OS")

	close({
		kind: "local"
		// No config for local - it represents the executing machine
	})
}
