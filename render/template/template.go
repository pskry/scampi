// SPDX-License-Identifier: GPL-3.0-only

package template

import (
	"strings"
	"text/template"

	"scampi.dev/scampi/errs"
)

// Renderable is the contract for anything the template renderer can render.
// Every caller supplies its own concrete type; the renderer never accepts
// bare strings.  Contract tested in test/template_render_test.go.
type Renderable interface {
	TemplateID() string
	TemplateText() string
	TemplateData() any
}

func Render(r Renderable) (string, bool) {
	name := r.TemplateID()
	tmpl := r.TemplateText()
	data := r.TemplateData()

	t, err := template.
		New(name).
		Option("missingkey=error").
		Funcs(template.FuncMap{
			"join": join,
		}).
		Parse(tmpl)
	if err != nil {
		panic(errs.BUG("template '%s' failed parsing: %w", tmpl, err))
	}

	b := strings.Builder{}
	// NOTE: at this point we MUST be able to trust that the template renders
	if err := t.Execute(&b, data); err != nil {
		panic(errs.BUG("template '%s' failed to render: %w", tmpl, err))
	}

	res := b.String()
	return res, strings.TrimSpace(res) != ""
}

// Template funcs
// -----------------------------------------------------------------------------

func join(sep string, s []string) string {
	return strings.Join(s, sep)
}
