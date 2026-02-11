package test

import "godoit.dev/doit/builtin"

targets: {
	local: builtin.local
}

deploy: {
	test: {
		targets: ["local"]
		steps: [
			builtin.template & {
				desc:    "already rendered"
				content: "hello"
				dest:    "/tmp/out.txt"
				perm:    "0644"
				owner:   "testuser"
				group:   "testgroup"
			},
		]
	}
}
