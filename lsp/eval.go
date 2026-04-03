// SPDX-License-Identifier: GPL-3.0-only

package lsp

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"

	"go.lsp.dev/protocol"
	"go.lsp.dev/uri"

	"scampi.dev/scampi/diagnostic"
	"scampi.dev/scampi/mod"
	rtmpl "scampi.dev/scampi/render/template"
	"scampi.dev/scampi/signal"
	"scampi.dev/scampi/source"
	"scampi.dev/scampi/spec"
	"scampi.dev/scampi/star"
	"scampi.dev/scampi/star/testkit"
)

// evaluate runs the full Starlark evaluation pipeline and returns LSP
// diagnostics. This catches everything the engine would catch: unknown
// kwargs, missing required fields, type errors, invalid enum values, etc.
func (s *Server) evaluate(ctx context.Context, docURI protocol.DocumentURI, content string) []protocol.Diagnostic {
	filePath := uriToPath(docURI)
	if filePath == "" {
		return nil
	}

	dir := filepath.Dir(filePath)
	base := source.LocalPosixSource{}
	src := source.WithRoot(dir, &overlaySource{
		base:    base,
		path:    filePath,
		content: []byte(content),
	})
	store := diagnostic.NewSourceStore()

	var baseOpts []star.EvalOption
	if strings.HasSuffix(filePath, "_test.scampi") {
		baseOpts = append(baseOpts, star.WithTestBuiltins(testkit.NewCollector()))
	}
	if m := s.moduleForFile(filePath); m != nil {
		baseOpts = append(baseOpts, star.WithModule(m, s.cacheDir))
	}

	// First pass: strict eval. If it succeeds, no diagnostics.
	_, err := star.Eval(ctx, filePath, store, src, baseOpts...)
	if err == nil {
		return nil
	}

	diags := evalErrors(err)

	// If the only errors are secret-related, re-eval with lenient
	// secrets to find structural errors beyond the secret failure.
	if allSecretErrors(diags) {
		lenientOpts := append(baseOpts, star.WithLenientSecrets())
		_, err = star.Eval(ctx, filePath, store, src, lenientOpts...)
		if err != nil {
			diags = append(diags, evalErrors(err)...)
		}
	}

	return diags
}

func allSecretErrors(diags []protocol.Diagnostic) bool {
	for _, d := range diags {
		if d.Severity != protocol.DiagnosticSeverityHint {
			return false
		}
	}
	return len(diags) > 0
}

// moduleForFile returns the parsed scampi.mod for the given file, walking
// up from the file's directory. Returns the server's cached module if the
// file is under the workspace root, or attempts to find one on disk.
func (s *Server) moduleForFile(filePath string) *mod.Module {
	if s.module != nil {
		modDir := filepath.Dir(s.module.Filename)
		if strings.HasPrefix(filePath, modDir+string(filepath.Separator)) {
			return s.module
		}
	}
	// Walk up from file looking for scampi.mod.
	dir := filepath.Dir(filePath)
	for {
		modPath := filepath.Join(dir, "scampi.mod")
		data, err := os.ReadFile(modPath)
		if err == nil {
			m, err := mod.Parse(modPath, data)
			if err == nil {
				return m
			}
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return nil
}

func evalErrors(err error) []protocol.Diagnostic {
	// Multi-diagnostic errors (multiple issues collected).
	var multi diagnostic.MultiDiagnostic
	if errors.As(err, &multi) {
		var diags []protocol.Diagnostic
		for _, d := range multi.Diagnostics() {
			diags = append(diags, diagnosticToLSP(d))
		}
		return diags
	}

	// Single diagnostic error.
	var diag diagnostic.Diagnostic
	if errors.As(err, &diag) {
		return []protocol.Diagnostic{diagnosticToLSP(diag)}
	}

	// Fallback: could be a syntax error from the Starlark parser.
	return syntaxErrors(err)
}

// secretErrorIDs are diagnostic IDs that should be downgraded to hints
// in the LSP. Secret decryption is a runtime concern — a config is
// structurally valid even if the editor can't decrypt secrets.
var secretErrorIDs = map[string]bool{
	"star.SecretBackendError": true,
	"star.SecretsConfigError": true,
}

func diagnosticToLSP(d diagnostic.Diagnostic) protocol.Diagnostic {
	tmpl := d.EventTemplate()

	var severity protocol.DiagnosticSeverity
	switch {
	case secretErrorIDs[tmpl.ID]:
		severity = protocol.DiagnosticSeverityHint
	case d.Severity() == signal.Warning:
		severity = protocol.DiagnosticSeverityWarning
	case d.Severity() == signal.Info || d.Severity() == signal.Notice:
		severity = protocol.DiagnosticSeverityInformation
	case d.Severity() == signal.Debug:
		severity = protocol.DiagnosticSeverityHint
	default:
		severity = protocol.DiagnosticSeverityError
	}

	msg, _ := rtmpl.Render(tmpl.TextField())
	if hint, ok := rtmpl.Render(tmpl.HintField()); ok {
		msg += "\nhint: " + hint
	}

	r := protocol.Range{}
	if tmpl.Source != nil {
		r = spanToRange(*tmpl.Source)
	}

	return protocol.Diagnostic{
		Range:    r,
		Severity: severity,
		Source:   "scampi",
		Message:  msg,
	}
}

func spanToRange(s spec.SourceSpan) protocol.Range {
	startLine := uint32(0)
	if s.StartLine > 0 {
		startLine = uint32(s.StartLine - 1)
	}
	startCol := uint32(0)
	if s.StartCol > 0 {
		startCol = uint32(s.StartCol - 1)
	}
	endLine := startLine
	if s.EndLine > 0 {
		endLine = uint32(s.EndLine - 1)
	}
	endCol := startCol
	if s.EndCol > 0 {
		endCol = uint32(s.EndCol - 1)
	}
	return protocol.Range{
		Start: protocol.Position{Line: startLine, Character: startCol},
		End:   protocol.Position{Line: endLine, Character: endCol},
	}
}

func uriToPath(u protocol.DocumentURI) string {
	return uri.URI(u).Filename()
}

// overlaySource wraps a real source.Source but returns in-memory content
// for a single file (the open document). Everything else passes through.
type overlaySource struct {
	base    source.Source
	path    string
	content []byte
}

func (o *overlaySource) ReadFile(ctx context.Context, path string) ([]byte, error) {
	abs, _ := filepath.Abs(path)
	if abs == o.path {
		return o.content, nil
	}
	return o.base.ReadFile(ctx, path)
}

func (o *overlaySource) WriteFile(ctx context.Context, path string, data []byte) error {
	return o.base.WriteFile(ctx, path, data)
}

func (o *overlaySource) EnsureDir(ctx context.Context, path string) error {
	return o.base.EnsureDir(ctx, path)
}

func (o *overlaySource) Stat(ctx context.Context, path string) (source.FileMeta, error) {
	return o.base.Stat(ctx, path)
}

func (o *overlaySource) LookupEnv(key string) (string, bool) {
	return o.base.LookupEnv(key)
}

func (o *overlaySource) LookupSecret(key string) (string, bool, error) {
	// Try real decryption first; if it fails, return a placeholder so
	// evaluation can continue past secret() calls. The LSP cares about
	// structural validity, not runtime secret access.
	val, ok, err := o.base.LookupSecret(key)
	if err == nil {
		return val, ok, nil
	}
	return "<secret:" + key + ">", true, nil
}
