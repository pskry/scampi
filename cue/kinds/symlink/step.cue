package symlink

#Step: close({
	_kind:  "symlink"
	desc?:  string
	target: string // path the symlink points to (like ln -s TARGET)
	link:   string // path where symlink is created (like ln -s ... LINK)
})
