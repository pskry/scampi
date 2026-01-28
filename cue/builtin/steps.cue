package builtin

import (
	kcopy "godoit.dev/doit/kinds/copy"
	ksymlink "godoit.dev/doit/kinds/symlink"
	ktemplate "godoit.dev/doit/kinds/template"
)

#BuiltinStep: kcopy.#Step | ksymlink.#Step | ktemplate.#Step
copy:         kcopy.#Step
symlink:      ksymlink.#Step
template:     ktemplate.#Step
