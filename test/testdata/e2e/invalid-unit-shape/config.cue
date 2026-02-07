package test

import "godoit.dev/doit/builtin"

targets: {
	local: builtin.local
}

// deploy block name must be valid identifier, testing invalid field shape
deploy: {
	test: {
		targets: ["local"]
		// Adding invalid extra field to test schema validation
		invalid_extra_field: "should not be allowed"
		steps: [
			builtin.copy & {
				desc:  "copy file"
				src:   "/src.txt"
				dest:  "/dest.txt"
				perm:  "0644"
				owner: "testuser"
				group: "testgroup"
			},
		]
	}
}
