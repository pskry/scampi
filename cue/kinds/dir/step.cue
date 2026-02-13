// SPDX-License-Identifier: GPL-3.0-only

package dir

#Step: {
	@doc("Ensure a directory exists with optional permissions and ownership")
	@example("""
		builtin.dir & {
		    path: "/opt/app/data"
		    perm: "0755"
		}
		""")

	close({
		kind:   "dir"
		desc?:  string @doc("Human-readable description")
		path:   string @doc("Absolute path to ensure exists (creates parents)")
		perm?:  string @doc("Permissions in octal notation (e.g. \"0755\")")
		owner?: string @doc("Owner user name or UID")
		group?: string @doc("Owner group name or GID")
	})
}
