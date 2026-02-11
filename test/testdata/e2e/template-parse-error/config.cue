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
				desc:    "bad syntax"
				content: "hello {{.name"
				dest:    "/tmp/out.txt"
				perm:    "0644"
				owner:   "testuser"
				group:   "testgroup"
			},
		]
	}
}
