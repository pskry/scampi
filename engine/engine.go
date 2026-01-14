package engine

import (
	"godoit.dev/doit/diagnostic"
	"godoit.dev/doit/source"
	"godoit.dev/doit/target"
)

type Engine struct {
	src source.Source
	tgt target.Target
	em  diagnostic.Emitter
}

func New(src source.Source, tgt target.Target, em diagnostic.Emitter) *Engine {
	return &Engine{
		src: src,
		tgt: tgt,
		em:  em,
	}
}
