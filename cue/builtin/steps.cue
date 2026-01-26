package builtin

import (
	kcopy "godoit.dev/doit/kinds/copy"
	ksymlink "godoit.dev/doit/kinds/symlink"
)

#BuiltinStep: kcopy.#Step | ksymlink.#Step
copy:         kcopy.#Step
symlink:      ksymlink.#Step
