// SPDX-License-Identifier: GPL-3.0-only

package lsp

import (
	"strings"

	"go.lsp.dev/protocol"

	"scampi.dev/scampi/lang/ast"
)

// funcLookupResult holds the result of a function lookup, including
// whether it was resolved via UFCS (in which case positional param
// indices are shifted by 1 since the receiver is implicit).
type funcLookupResult struct {
	Func FuncInfo
	UFCS bool
}

// lookupFunc checks the stdlib catalog first, then falls back to resolving
// user-defined functions from the current file. When the name is dotted
// (e.g. `x.yo` from a UFCS-style call or selector), the last segment is
// also tried so the function-name part of `x.yo()` resolves to `yo`.
//
// For UFCS calls like `resolver.get(...)`, the cursor reports `resolver.get`
// but the catalog has `secrets.get`. When direct and tail lookups both miss,
// we try every known module prefix with the tail segment.
func (s *Server) lookupFunc(docURI protocol.DocumentURI, name string) (FuncInfo, bool) {
	r := s.lookupFuncEx(docURI, name)
	return r.Func, r.Func.Name != ""
}

func (s *Server) lookupFuncEx(docURI protocol.DocumentURI, name string) funcLookupResult {
	if f, ok := s.catalog.Lookup(name); ok {
		return funcLookupResult{Func: f}
	}
	if f, ok := s.resolveUserFunc(docURI, name); ok {
		return funcLookupResult{Func: f}
	}
	if i := strings.LastIndexByte(name, '.'); i >= 0 && i < len(name)-1 {
		tail := name[i+1:]
		if f, ok := s.catalog.Lookup(tail); ok {
			return funcLookupResult{Func: f}
		}
		if f, ok := s.resolveUserFunc(docURI, tail); ok {
			return funcLookupResult{Func: f}
		}
		// UFCS: `var.method` — try every module as prefix.
		for _, mod := range s.catalog.Modules() {
			if f, ok := s.catalog.Lookup(mod + "." + tail); ok {
				return funcLookupResult{Func: f, UFCS: true}
			}
		}
	}
	return funcLookupResult{}
}

// resolveUserFunc attempts to find a user-defined function by name in the
// current file. Returns a FuncInfo with params extracted from the
// FuncDecl, or false if not found.
func (s *Server) resolveUserFunc(docURI protocol.DocumentURI, name string) (FuncInfo, bool) {
	doc, ok := s.docs.Get(docURI)
	if !ok {
		return FuncInfo{}, false
	}

	filePath := uriToPath(docURI)
	f, _ := Parse(filePath, []byte(doc.Content))
	if f == nil {
		return FuncInfo{}, false
	}

	if bf, ok := funcDeclToLSP(f, name); ok {
		return bf, true
	}

	return FuncInfo{}, false
}

// funcDeclToLSP finds a FuncDecl by name and converts it to FuncInfo.
func funcDeclToLSP(f *ast.File, name string) (FuncInfo, bool) {
	for _, d := range f.Decls {
		fd, ok := d.(*ast.FuncDecl)
		if !ok || fd.Name.Name != name {
			continue
		}

		params := fieldsToParams(fd.Params, "")
		return FuncInfo{
			Name:   name,
			Params: params,
		}, true
	}
	return FuncInfo{}, false
}
