package sandbox

import "godoit.dev/doit/builtin"

targets: {
	local: builtin.local
}

deploy: {
	test: {
		targets: ["local"]
		steps: [
			builtin.symlink & {
				desc: "symlink missing target field"
				link: "/tmp/mylink"
			},
		]
	}
}
