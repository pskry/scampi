package test

import "godoit.dev/doit/builtin"

targets: {
	local: builtin.local
}

deploy: {
	test: {
		targets: ["local"]
		steps: [
			builtin.dir & {
				desc: "directory already exists"
				path: "/tmp/mydir"
			},
		]
	}
}
