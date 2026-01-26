package test

import "godoit.dev/doit/builtin"

steps: [
	builtin.symlink & {
		desc:   "update symlink target"
		target: "/new-target.txt"
		link:   "/link.txt"
	},
]
