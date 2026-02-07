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
				desc:  "copy without perm"
				src:   "/src.txt"
				dest:  "/dest.txt"
				owner: "testuser"
				group: "testgroup"
			},
		]
	}
}
