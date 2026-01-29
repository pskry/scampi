package copy

import (
	"bytes"
	"context"
	"fmt"
	"path/filepath"

	"godoit.dev/doit/capability"
	"godoit.dev/doit/diagnostic"
	"godoit.dev/doit/diagnostic/event"
	"godoit.dev/doit/signal"
	"godoit.dev/doit/source"
	"godoit.dev/doit/spec"
	"godoit.dev/doit/step/sharedops"
	"godoit.dev/doit/target"
)

const copyFileID = "builtin.copy-file"

type copyFileOp struct {
	sharedops.BaseOp
	src  string
	dest string
}

func (op *copyFileOp) Check(ctx context.Context, src source.Source, tgt target.Target) (spec.CheckResult, error) {
	fsTgt := target.Must[target.Filesystem](copyFileID, tgt)

	// source must exist
	srcData, err := src.ReadFile(ctx, op.src)
	if err != nil {
		return spec.CheckUnsatisfied, CopySourceMissing{
			Path:   op.src,
			Err:    err,
			Source: op.SrcSpan,
		}
	}

	// destination parent must exist
	if _, err := fsTgt.Stat(ctx, filepath.Dir(op.dest)); err != nil {
		return spec.CheckUnsatisfied, CopyDestDirMissing{
			Path:   filepath.Dir(op.dest),
			Err:    err,
			Source: op.DestSpan,
		}
	}

	// dest file comparison (expected drift)
	destData, err := fsTgt.ReadFile(ctx, op.dest)
	if err != nil {
		return spec.CheckUnsatisfied, nil
	}

	if !bytes.Equal(srcData, destData) {
		return spec.CheckUnsatisfied, nil
	}

	return spec.CheckSatisfied, nil
}

func (op *copyFileOp) Execute(ctx context.Context, src source.Source, tgt target.Target) (spec.Result, error) {
	fsTgt := target.Must[target.Filesystem](copyFileID, tgt)

	srcData, err := src.ReadFile(ctx, op.src)
	if err != nil {
		return spec.Result{}, err
	}

	destData, err := fsTgt.ReadFile(ctx, op.dest)
	if err == nil && bytes.Equal(srcData, destData) {
		return spec.Result{Changed: false}, nil
	}

	if err := fsTgt.WriteFile(ctx, op.dest, srcData); err != nil {
		return spec.Result{}, err
	}

	return spec.Result{Changed: true}, nil
}

func (copyFileOp) RequiredCapabilities() capability.Capability {
	return capability.Filesystem
}

type copyFileDesc struct {
	Src  string
	Dest string
}

func (d copyFileDesc) PlanTemplate() spec.PlanTemplate {
	return spec.PlanTemplate{
		ID:   copyFileID,
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

type CopySourceMissing struct {
	Path   string
	Source spec.SourceSpan
	Err    error
}

func (e CopySourceMissing) Error() string {
	return fmt.Sprintf("source file %q does not exist", e.Path)
}

func (e CopySourceMissing) EventTemplate() event.Template {
	return event.Template{
		ID:     "builtin.copy.SourceMissing",
		Text:   `source file "{{.Path}}" does not exist`,
		Hint:   "ensure the source file exists and is readable",
		Help:   "the copy action cannot proceed without a readable source file",
		Data:   e,
		Source: &e.Source,
	}
}

func (CopySourceMissing) Severity() signal.Severity { return signal.Error }
func (CopySourceMissing) Impact() diagnostic.Impact { return diagnostic.ImpactAbort }

type CopyDestDirMissing struct {
	Path   string
	Source spec.SourceSpan
	Err    error
}

func (e CopyDestDirMissing) Error() string {
	return fmt.Sprintf("destination directory %q does not exist", e.Path)
}

func (e CopyDestDirMissing) EventTemplate() event.Template {
	return event.Template{
		ID:     "builtin.copy.DestDirMissing",
		Text:   `destination directory "{{.Path}}" does not exist`,
		Hint:   "create the destination directory before running this action",
		Help:   "the copy action does not create directories automatically",
		Data:   e,
		Source: &e.Source,
	}
}

func (CopyDestDirMissing) Severity() signal.Severity { return signal.Error }
func (CopyDestDirMissing) Impact() diagnostic.Impact { return diagnostic.ImpactAbort }
