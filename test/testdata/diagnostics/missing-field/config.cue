package sandbox

import "godoit.dev/doit/builtin"

targets: {
	local: builtin.local
}

deploy: {
	test: {
		targets: ["local"]
		steps: [
			builtin.copy & {
				desc:  "builtin.copy action"
				src:   "./.src1.yml"
				dest:  "./.dest1.yml"
				owner: "user"
				group: "group"
			},
			{
				kind:  "copy"
				desc:  "anon copy action"
				src:   "./.src1.yml"
				dest:  "./.dest1.yml"
				perm:  "0644"
				owner: "user"
				group: "group"
			},
		]
	}
}
