package test

import "godoit.dev/doit/builtin"

targets: {
	local: builtin.local
}

deploy: {
	test: {
		targets: ["local"]
		steps: [
			builtin.symlink & {
				desc:   "create symlink"
				target: "/tmp/target.txt"
				link:   "/tmp/link.txt"
			},
		]
	}
}
