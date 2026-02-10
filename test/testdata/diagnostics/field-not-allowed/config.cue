package sandbox

import "godoit.dev/doit/builtin"

targets: {
	local: builtin.local
}

deploy: {
	test: {
		targets: ["local"]
		extra_field: "not allowed"
		steps: [
			builtin.copy & {
				desc:  "copy file"
				src:   "/src.txt"
				dest:  "/dest.txt"
				perm:  "0644"
				owner: "root"
				group: "root"
			},
		]
	}
}
