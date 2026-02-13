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
				desc: "create new directory"
				path: "/tmp/mydir"
			},
		]
	}
}
