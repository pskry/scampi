package test

import "godoit.dev/doit/builtin"

targets: {
	local: builtin.local
}

deploy: {
	test: {
		targets: ["local"]
		steps: [
			builtin.copy & {
				desc:  "copy with bad permission"
				src:   "/src.txt"
				dest:  "/dest.txt"
				perm:  "invalid"
				owner: "testuser"
				group: "testgroup"
			},
		]
	}
}
