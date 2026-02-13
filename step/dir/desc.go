// SPDX-License-Identifier: GPL-3.0-only

package dir

import (
	"godoit.dev/doit/spec"
)

type ensureDirDesc struct {
	Path string
}

func (d ensureDirDesc) PlanTemplate() spec.PlanTemplate {
	return spec.PlanTemplate{
		ID:   id,
		Text: `ensure directory "{{.Path}}"`,
		Data: d,
	}
}

func (op *ensureDirOp) OpDescription() spec.OpDescription {
	return ensureDirDesc{
		Path: op.path,
	}
}
