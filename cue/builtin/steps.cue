// SPDX-License-Identifier: GPL-3.0-only

package builtin

import (
	kcopy "godoit.dev/doit/kinds/copy"
	kdir "godoit.dev/doit/kinds/dir"
	kpkg "godoit.dev/doit/kinds/pkg"
	ksymlink "godoit.dev/doit/kinds/symlink"
	ktemplate "godoit.dev/doit/kinds/template"
)

#BuiltinStep: kcopy.#Step | kdir.#Step | kpkg.#Step | ksymlink.#Step | ktemplate.#Step
copy:         kcopy.#Step
dir:          kdir.#Step
pkg:          kpkg.#Step
symlink:      ksymlink.#Step
template:     ktemplate.#Step
