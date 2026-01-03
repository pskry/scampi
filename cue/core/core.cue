package core

{
	tasks: #TaskMap
}

#TaskMap: {
	[string]: #Task
}

#Task: {
	meta: {
		kind: string
	}

	...
}
