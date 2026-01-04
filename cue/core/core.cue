package core

close({
	tasks: #TaskMap
})

#TaskMap: {
	[string]: #Task
}

#Task: {
	meta: {
		kind: string
	}

	...
}
