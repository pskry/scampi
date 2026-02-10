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
				desc:  "copy with relative dest"
				src:   "/tmp/src.txt"
				dest:  "./relative/dest.txt"
				perm:  "0644"
				owner: "user"
				group: "group"
			},
		]
	}
}
