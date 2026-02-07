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
				desc:   "link in missing directory"
				target: "/tmp/target.txt"
				link:   "/tmp/nonexistent/link.txt"
			},
		]
	}
}
