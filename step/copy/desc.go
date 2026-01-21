package copy

import (
	"godoit.dev/doit/spec"
)

type copyFileDesc struct {
	Src  string
	Dest string
}

func (d copyFileDesc) PlanTemplate() spec.PlanTemplate {
	return spec.PlanTemplate{
		ID:   "plan.copy.file",
		Text: `copy "{{.Src}}" -> "{{.Dest}}"`,
		Data: d,
	}
}

func (op *copyFileOp) OpDescription() spec.OpDescription {
	return copyFileDesc{
		Src:  op.src,
		Dest: op.dest,
	}
}

type ensureOwnerDesc struct {
	User  string
	Group string
	Path  string
}

func (d ensureOwnerDesc) PlanTemplate() spec.PlanTemplate {
	return spec.PlanTemplate{
		ID:   "plan.copy.ensureOwner",
		Text: `ensure owner "{{.User}}:{{.Group}}" on "{{.Path}}"`,
		Data: d,
	}
}

func (op *ensureOwnerOp) OpDescription() spec.OpDescription {
	return ensureOwnerDesc{
		User:  op.owner,
		Group: op.group,
		Path:  op.path,
	}
}

type ensureModeDesc struct {
	Mode string
	Path string
}

func (d ensureModeDesc) PlanTemplate() spec.PlanTemplate {
	return spec.PlanTemplate{
		ID:   "plan.copy.ensureMode",
		Text: `ensure mode {{.Mode}} on "{{.Path}}"`,
		Data: d,
	}
}

func (op *ensureModeOp) OpDescription() spec.OpDescription {
	return ensureModeDesc{
		Mode: op.mode.String(),
		Path: op.path,
	}
}
