package test

import "godoit.dev/doit/builtin"

steps: [
	builtin.copy & {
		desc:  "copy missing file"
		src:   "/nonexistent.txt"
		dest:  "/out.txt"
		perm:  "0644"
		owner: "testuser"
		group: "testgroup"
	},
]
