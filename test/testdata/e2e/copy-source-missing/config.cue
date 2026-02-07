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
				desc:  "copy missing file"
				src:   "/tmp/nonexistent.txt"
				dest:  "/tmp/out.txt"
				perm:  "0644"
				owner: "testuser"
				group: "testgroup"
			},
		]
	}
}
