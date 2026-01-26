package test

import "godoit.dev/doit/builtin"

steps: [
	builtin.symlink & {
		desc:   "symlink already correct"
		target: "/target.txt"
		link:   "/link.txt"
	},
]
