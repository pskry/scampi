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
				desc: "create directory with permissions"
				path: "/tmp/mydir"
				perm: "0700"
			},
		]
	}
}
