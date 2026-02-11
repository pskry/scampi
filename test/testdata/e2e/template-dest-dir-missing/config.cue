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
				desc:    "missing parent dir"
				content: "hello"
				dest:    "/no/such/dir/out.txt"
				perm:    "0644"
				owner:   "testuser"
				group:   "testgroup"
			},
		]
	}
}
