package core

playbook: {
	tasks: [...#Task]
}

#Task: {
	kind: string
	...
}
