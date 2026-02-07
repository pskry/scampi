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
				desc:   "link path is a regular file"
				target: "/tmp/target.txt"
				link:   "/tmp/link.txt"
			},
		]
	}
}
