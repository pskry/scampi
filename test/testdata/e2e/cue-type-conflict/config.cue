package test

import "godoit.dev/doit/builtin"

targets: {
	local: builtin.local
}

// src must be a string, not a number
deploy: {
	test: {
		targets: ["local"]
		steps: [
			builtin.copy & {
				desc:  "copy file"
				src:   123
				dest:  "/dest.txt"
				perm:  "0644"
				owner: "testuser"
				group: "testgroup"
			},
		]
	}
}
