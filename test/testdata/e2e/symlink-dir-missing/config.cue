package test

import "godoit.dev/doit/builtin"

steps: [
	builtin.symlink & {
		desc:   "link in missing directory"
		target: "/target.txt"
		link:   "/nonexistent/link.txt"
	},
]
