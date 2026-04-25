// SPDX-License-Identifier: GPL-3.0-only

package lsp

import (
	"testing"

	"go.lsp.dev/protocol"
)

func TestCompletion_UserType_InList_NewLine(t *testing.T) {
	s := testServer()
	docURI := protocol.DocumentURI("file:///test.scampi")
	// Same scenario but user pressed ENTER after "Container {" and
	// is on a new indented line.
	text := `module main

type Container {
  id:       int
  hostname: string
}

let items = [
  Container { id = 1, hostname = "a" },
  Container {

  },
]
`
	s.docs.Open(docURI, text, 1)

	// Cursor on the empty indented line inside the second Container
	items := completionAt(t, s, docURI, 10, 4)
	if len(items) == 0 {
		t.Fatal("expected field completions on new line inside user type in list")
	}
	requireLabels(t, items, "id", "hostname")
}

func TestCompletion_UserType_InList_PartialFields(t *testing.T) {
	s := testServer()
	docURI := protocol.DocumentURI("file:///test.scampi")
	text := `module main

type Server {
  name: string
  port: int
  tls:  bool = false
}

let servers = [
  Server { name = "web", port = 443 },
  Server { name = "api",
  },
]
`
	s.docs.Open(docURI, text, 1)

	// Cursor after "name = \"api\"," — should offer port and tls but not name.
	items := completionAt(t, s, docURI, 10, 25)
	if len(items) == 0 {
		t.Fatal("expected completions")
	}
	requireLabels(t, items, "port", "tls")
	rejectLabels(t, items, "name")
}
