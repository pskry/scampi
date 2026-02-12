// SPDX-License-Identifier: GPL-3.0-only

package pkg

#Step: {
	@doc("Ensure packages are present, absent, or at the latest version on the target")
	@example("""
		builtin.pkg & {
		    packages: ["nginx", "curl", "git"]
		}
		""")
	@example("""
		builtin.pkg & {
		    packages: ["telnetd"]
		    state:    "absent"
		}
		""")
	@example("""
		builtin.pkg & {
		    packages: ["nginx"]
		    state:    "latest"
		}
		""")

	close({
		kind:  "pkg"
		desc?: string @doc("Human-readable description")
		packages: [...string] & [_, ...] @doc("Packages to manage")
		state: *"present" | "absent" | "latest" @doc("Desired package state")
	})
}
