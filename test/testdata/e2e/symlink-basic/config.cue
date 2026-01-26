package test

import "godoit.dev/doit/builtin"

steps: [
	builtin.symlink & {
		desc:   "create symlink"
		target: "/target.txt"
		link:   "/link.txt"
	},
]
