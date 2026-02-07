package test

import "godoit.dev/doit/builtin"

targets: {
	local: builtin.local
}

// Reference to undefined variable
deploy: {
	test: {
		targets: ["local"]
		steps: [
			builtin.copy & {
				desc:  "copy file"
				src:   undefinedVar
				dest:  "/dest.txt"
				perm:  "0644"
				owner: "testuser"
				group: "testgroup"
			},
		]
	}
}
