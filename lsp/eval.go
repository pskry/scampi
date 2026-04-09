// SPDX-License-Identifier: GPL-3.0-only

package lsp

import (
	"context"
	"os"
	"path/filepath"

	"go.lsp.dev/protocol"
	"go.lsp.dev/uri"

	"scampi.dev/scampi/lang/ast"
	"scampi.dev/scampi/lang/check"
	"scampi.dev/scampi/lang/lex"
	"scampi.dev/scampi/lang/parse"
	"scampi.dev/scampi/lang/token"
	"scampi.dev/scampi/mod"
	"scampi.dev/scampi/std"
)

// bootstrapModules loads the standard library stubs once so the type
// checker can resolve imports in user files.
func bootstrapModules() map[string]*check.Scope {
	modules, err := check.BootstrapModules(std.FS)
	if err != nil {
		// Stubs are compiled-in; failure here is a build bug.
		panic("lsp: failed to bootstrap stdlib: " + err.Error())
	}
	return modules
}

// loadUserModules parses and type-checks user module dependencies from
// scampi.mod, adding their scopes to the module map so the checker can
// resolve imports in user code.
func (s *Server) loadUserModules() {
	if s.module == nil {
		return
	}
	for _, dep := range s.module.Require {
		dir := depDir(s.module, &dep)
		data, path := readModuleEntry(dir, lastPathSegment(dep.Path))
		if data == nil {
			s.log.Printf("user module %s: no entry point in %s", dep.Path, dir)
			continue
		}

		l := lex.New(path, data)
		p := parse.New(l)
		f := p.Parse()
		if f == nil || f.Module == nil {
			s.log.Printf("user module %s: parse failed", dep.Path)
			continue
		}

		c := check.New(s.modules)
		c.Check(f)
		modName := f.Module.Name.Name
		s.modules[modName] = c.FileScope()

		// Register funcs/decls into catalog and goto-def index.
		s.registerModuleEntries(f, modName, path, data)
		s.log.Printf("user module %s: loaded as %q", dep.Path, modName)
	}
}

// registerModuleEntries adds a user module's funcs and decls to the
// catalog (for hover/completion) and stubDefs (for goto-def).
func (s *Server) registerModuleEntries(f *ast.File, modName, filePath string, src []byte) {
	for _, d := range f.Decls {
		switch d := d.(type) {
		case *ast.FuncDecl:
			name := modName + "." + d.Name.Name
			info := funcDeclToInfo(d, modName)
			info.Name = name
			s.catalog.funcs[name] = info
			s.stubDefs.locs[name] = stubLocation{
				path: filePath, src: src, span: d.Name.SrcSpan,
			}
		case *ast.DeclDecl:
			dn := declName(d)
			name := modName + "." + dn
			info := declDeclToInfo(d, modName)
			info.Name = name
			s.catalog.funcs[name] = info
			s.stubDefs.locs[name] = stubLocation{
				path: filePath, src: src, span: d.Name.SrcSpan,
			}
		}
	}
	// Rebuild the catalog index so new entries show up in completion.
	s.catalog.buildIndex()
}

func depDir(m *mod.Module, dep *mod.Dependency) string {
	if dep.IsLocal() {
		dir := dep.Version
		if !filepath.IsAbs(dir) {
			dir = filepath.Join(filepath.Dir(m.Filename), dir)
		}
		return dir
	}
	return filepath.Join(mod.DefaultCacheDir(), dep.Path+"@"+dep.Version)
}

func lastPathSegment(p string) string {
	if i := len(p) - 1; i >= 0 {
		for ; i >= 0; i-- {
			if p[i] == '/' {
				return p[i+1:]
			}
		}
	}
	return p
}

// readModuleEntry finds the entry point .scampi file in a module
// directory, trying _index.scampi then <name>.scampi.
func readModuleEntry(dir, name string) ([]byte, string) {
	for _, candidate := range []string{
		filepath.Join(dir, "_index.scampi"),
		filepath.Join(dir, name+".scampi"),
	} {
		data, err := os.ReadFile(candidate)
		if err == nil {
			return data, candidate
		}
	}
	return nil, ""
}

// evaluate runs the scampi-lang lex → parse → check pipeline and
// returns LSP diagnostics.
func (s *Server) evaluate(_ context.Context, docURI protocol.DocumentURI, content string) []protocol.Diagnostic {
	filePath := uriToPath(docURI)
	if filePath == "" {
		return nil
	}

	data := []byte(content)

	// Parse.
	f, parseDiags := Parse(filePath, data)
	if len(parseDiags) > 0 {
		return parseDiags
	}
	if f == nil {
		return nil
	}

	// Type check.
	c := check.New(s.modules)
	c.Check(f)

	var diags []protocol.Diagnostic
	for _, e := range c.Errors() {
		diags = append(diags, checkerErrorToLSP(data, e))
	}
	return diags
}

func checkerErrorToLSP(src []byte, e check.Error) protocol.Diagnostic {
	return spanDiag(src, e.Span, e.Msg)
}

func uriToPath(u protocol.DocumentURI) string {
	return uri.URI(u).Filename()
}

// spanToRange converts a token.Span to an LSP range, resolving byte
// offsets to line/column via the source bytes.
func tokenSpanToRange(src []byte, s token.Span) protocol.Range {
	start, end := token.ResolveSpan(src, s)
	return protocol.Range{
		Start: protocol.Position{
			Line:      uint32(max(start.Line-1, 0)),
			Character: uint32(max(start.Col-1, 0)),
		},
		End: protocol.Position{
			Line:      uint32(max(end.Line-1, 0)),
			Character: uint32(max(end.Col-1, 0)),
		},
	}
}

// diagnoseFile reads and diagnoses a single file on disk.
func (s *Server) diagnoseFile(ctx context.Context, path string) {
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}
	docURI := protocol.DocumentURI(uri.File(path))
	diags := s.evaluate(ctx, docURI, string(data))
	if diags == nil {
		diags = []protocol.Diagnostic{}
	}
	s.log.Printf("workspace diag: %s → %d", path, len(diags))
	_ = s.client.PublishDiagnostics(ctx, &protocol.PublishDiagnosticsParams{
		URI:         docURI,
		Diagnostics: diags,
	})
}
