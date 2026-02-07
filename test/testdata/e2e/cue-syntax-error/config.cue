package test

// Missing closing brace - syntax error
deploy: {
	test: {
		targets: ["local"]
		steps: [
			{
				desc: "broken"

