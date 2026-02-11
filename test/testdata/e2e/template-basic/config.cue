package test

import "godoit.dev/doit/builtin"

targets: {
	local: builtin.local
}

deploy: {
	test: {
		targets: ["local"]
		steps: [
			builtin.template & {
				desc: "render greeting"
				src:  "/tmpl/greeting.txt"
				dest: "/tmp/greeting.txt"
				data: {
					values: {
						name:  "world"
						count: 3
					}
				}
				perm:  "0644"
				owner: "testuser"
				group: "testgroup"
			},
		]
	}
}
