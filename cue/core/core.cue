// SPDX-License-Identifier: GPL-3.0-only

package core

import "godoit.dev/doit/builtin"

// targets defines available execution environments by name
targets: {
	[string]: builtin.#BuiltinTarget
}

// deploy defines what runs where
deploy: {
	[string]: #DeployBlock
}

#DeployBlock: {
	// targets lists target names from the targets map
	targets: [...string]

	// steps defines the ordered sequence of operations
	steps: [...builtin.#BuiltinStep]
}
