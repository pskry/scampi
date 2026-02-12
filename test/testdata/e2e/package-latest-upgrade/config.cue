package test

import "godoit.dev/doit/builtin"

targets: {
	local: builtin.local
}

deploy: {
	test: {
		targets: ["local"]
		steps: [
			builtin.pkg & {
				desc:     "upgrade packages"
				packages: ["nginx"]
				state:    "latest"
			},
		]
	}
}
