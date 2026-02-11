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
				desc:    "render inline"
				content: "server={{ .host }}:{{ .port }}"
				dest:    "/tmp/config.txt"
				data: {
					values: {
						host: "localhost"
						port: 8080
					}
				}
				perm:  "0600"
				owner: "testuser"
				group: "testgroup"
			},
		]
	}
}
