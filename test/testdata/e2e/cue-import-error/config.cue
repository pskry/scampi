package test

import "godoit.dev/doit/nonexistent"
import "godoit.dev/doit/builtin"

targets: {
	local: builtin.local
}

deploy: {
	test: {
		targets: ["local"]
		steps: [
			nonexistent.step & {
				desc: "broken"
			},
		]
	}
}
