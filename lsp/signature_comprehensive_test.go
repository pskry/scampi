// SPDX-License-Identifier: GPL-3.0-only

package lsp

import (
	"context"
	"testing"

	"go.lsp.dev/protocol"
)

func signatureAt(t *testing.T, s *Server, docURI protocol.DocumentURI, line, col uint32) *protocol.SignatureHelp {
	t.Helper()
	result, err := s.SignatureHelp(context.Background(), &protocol.SignatureHelpParams{
		TextDocumentPositionParams: protocol.TextDocumentPositionParams{
			TextDocument: protocol.TextDocumentIdentifier{URI: docURI},
			Position:     protocol.Position{Line: line, Character: col},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	return result
}

func TestSignature_StdlibFunc(t *testing.T) {
	s := testServer()
	docURI := protocol.DocumentURI("file:///test.scampi")
	text := `posix.copy(`
	s.docs.Open(docURI, text, 1)

	result := signatureAt(t, s, docURI, 0, uint32(len(text)))
	if result == nil {
		t.Fatal("expected signature help")
	}
	if len(result.Signatures) == 0 {
		t.Fatal("expected at least one signature")
	}
	if result.Signatures[0].Label == "" {
		t.Error("signature label should not be empty")
	}
}

func TestSignature_StructLitBraces(t *testing.T) {
	s := testServer()
	docURI := protocol.DocumentURI("file:///test.scampi")
	text := `posix.copy {`
	s.docs.Open(docURI, text, 1)

	result := signatureAt(t, s, docURI, 0, uint32(len(text)))
	if result == nil {
		t.Fatal("expected signature help inside struct literal braces")
	}
	if len(result.Signatures) == 0 {
		t.Fatal("expected at least one signature")
	}
}

func TestSignature_UserFunc(t *testing.T) {
	s := testServer()
	docURI := protocol.DocumentURI("file:///test.scampi")
	text := `module main

func greet(name: string, loud: bool = false) string {
  return "hi"
}

greet(
`
	s.docs.Open(docURI, text, 1)

	result := signatureAt(t, s, docURI, 6, 6)
	if result == nil {
		t.Fatal("expected signature help for user func")
	}
	if len(result.Signatures) == 0 {
		t.Fatal("expected signature")
	}
	sig := result.Signatures[0]
	if len(sig.Parameters) < 2 {
		t.Errorf("expected 2 params, got %d", len(sig.Parameters))
	}
}

func TestSignature_OutsideCall(t *testing.T) {
	s := testServer()
	docURI := protocol.DocumentURI("file:///test.scampi")
	text := `let x = 42`
	s.docs.Open(docURI, text, 1)

	result := signatureAt(t, s, docURI, 0, 10)
	if result != nil && len(result.Signatures) > 0 {
		t.Error("should not have signature help outside a call")
	}
}

func TestSignature_ActiveParam(t *testing.T) {
	s := testServer()
	docURI := protocol.DocumentURI("file:///test.scampi")
	text := `ssh.target(name = "web", host = "1.2.3.4", `
	s.docs.Open(docURI, text, 1)

	result := signatureAt(t, s, docURI, 0, uint32(len(text)))
	if result == nil {
		t.Fatal("expected signature help")
	}
	// After 2 commas, active param should be 2.
	if result.ActiveParameter != 2 {
		t.Errorf("active param = %d, want 2", result.ActiveParameter)
	}
}
