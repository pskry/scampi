// SPDX-License-Identifier: GPL-3.0-only

package lsp

import (
	"context"
	"strings"
	"testing"

	"go.lsp.dev/protocol"
)

func TestSignatureHelp(t *testing.T) {
	s := testServer()
	uri := protocol.DocumentURI("file:///test.scampi")
	text := `copy(`
	s.docs.Open(uri, text, 1)

	result, err := s.SignatureHelp(context.Background(), &protocol.SignatureHelpParams{
		TextDocumentPositionParams: protocol.TextDocumentPositionParams{
			TextDocument: protocol.TextDocumentIdentifier{URI: uri},
			Position:     protocol.Position{Line: 0, Character: uint32(len(text))},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if result == nil || len(result.Signatures) == 0 {
		t.Fatal("expected signature help")
	}

	sig := result.Signatures[0]
	if !strings.HasPrefix(sig.Label, "copy(") {
		t.Errorf("expected signature starting with 'copy(', got %q", sig.Label)
	}
	if len(sig.Parameters) == 0 {
		t.Error("expected parameters")
	}
}

func TestSignatureHelpActiveParam(t *testing.T) {
	s := testServer()
	uri := protocol.DocumentURI("file:///test.scampi")
	text := `copy(src=local("./f"), `
	s.docs.Open(uri, text, 1)

	result, err := s.SignatureHelp(context.Background(), &protocol.SignatureHelpParams{
		TextDocumentPositionParams: protocol.TextDocumentPositionParams{
			TextDocument: protocol.TextDocumentIdentifier{URI: uri},
			Position:     protocol.Position{Line: 0, Character: uint32(len(text))},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if result == nil {
		t.Fatal("expected signature help")
	}
	if result.ActiveParameter != 1 {
		t.Errorf("active param = %d, want 1", result.ActiveParameter)
	}
}

func TestSignatureHelpOutsideCall(t *testing.T) {
	s := testServer()
	uri := protocol.DocumentURI("file:///test.scampi")
	s.docs.Open(uri, "x = 1", 1)

	result, err := s.SignatureHelp(context.Background(), &protocol.SignatureHelpParams{
		TextDocumentPositionParams: protocol.TextDocumentPositionParams{
			TextDocument: protocol.TextDocumentIdentifier{URI: uri},
			Position:     protocol.Position{Line: 0, Character: 5},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if result != nil {
		t.Error("expected nil signature help outside a call")
	}
}
