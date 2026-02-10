package sandbox

import "godoit.dev/doit/builtin"

targets: {
	local: builtin.local
}

deploy: {
	test: {
		targets: ["local"]
		steps: [
			builtin.copy & {
				desc:  "copy with invalid perm"
				src:   "/tmp/src.txt"
				dest:  "/tmp/dest.txt"
				perm:  "bad"
				owner: "root"
				group: "root"
			},
		]
	}
}
