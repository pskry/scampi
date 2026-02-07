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
				desc:  "copy to nonexistent dir"
				src:   "/tmp/src.txt"
				dest:  "/tmp/subdir/dest.txt"
				perm:  "0644"
				owner: "testuser"
				group: "testgroup"
			},
		]
	}
}
