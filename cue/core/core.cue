package core

import "godoit.dev/doit/builtin"

unit?: close({
	id!:   string
	desc?: string
})

steps: [...close(builtin.#BuiltinStep)]
