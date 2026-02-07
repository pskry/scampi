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
				desc:   "update symlink target"
				target: "/tmp/new-target.txt"
				link:   "/tmp/link.txt"
			},
		]
	}
}
