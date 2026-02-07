package test

import "godoit.dev/doit/builtin"

targets: {
	local: builtin.local
}

// deploy.test.steps must be a list, not a struct
deploy: {
	test: {
		targets: ["local"]
		steps: {}
	}
}
