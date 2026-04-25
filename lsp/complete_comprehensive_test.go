// SPDX-License-Identifier: GPL-3.0-only

package lsp

import (
	"context"
	"strings"
	"testing"

	"go.lsp.dev/protocol"
)

// completionAt is a test helper that returns completion labels at the
// given line/col in a document.
func completionAt(t *testing.T, s *Server, docURI protocol.DocumentURI, line, col uint32) []protocol.CompletionItem {
	t.Helper()
	result, err := s.Completion(context.Background(), &protocol.CompletionParams{
		TextDocumentPositionParams: protocol.TextDocumentPositionParams{
			TextDocument: protocol.TextDocumentIdentifier{URI: docURI},
			Position:     protocol.Position{Line: line, Character: col},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if result == nil {
		return nil
	}
	return result.Items
}

func labels(items []protocol.CompletionItem) map[string]bool {
	m := make(map[string]bool, len(items))
	for _, item := range items {
		m[item.Label] = true
	}
	return m
}

func requireLabels(t *testing.T, items []protocol.CompletionItem, want ...string) {
	t.Helper()
	have := labels(items)
	for _, w := range want {
		if !have[w] {
			var all []string
			for _, item := range items {
				all = append(all, item.Label)
			}
			t.Errorf("expected %q in completions, got %v", w, all)
		}
	}
}

func rejectLabels(t *testing.T, items []protocol.CompletionItem, reject ...string) {
	t.Helper()
	have := labels(items)
	for _, r := range reject {
		if have[r] {
			t.Errorf("%q should NOT be in completions", r)
		}
	}
}

// User type kwargs in different positions
// -----------------------------------------------------------------------------

func TestCompletion_UserType_InListAfterExisting(t *testing.T) {
	s := testServer()
	docURI := protocol.DocumentURI("file:///test.scampi")
	text := `module main

type Item {
  id:   int
  name: string
  tag:  string = "default"
}

let items = [
  Item { id = 1, name = "a" },
  Item { id = 2,
]
`
	s.docs.Open(docURI, text, 1)

	items := completionAt(t, s, docURI, 10, 16)
	if len(items) == 0 {
		t.Fatal("expected field completions")
	}
	requireLabels(t, items, "name", "tag")
	rejectLabels(t, items, "id")
}

func TestCompletion_UserType_InNestedCall(t *testing.T) {
	s := testServer()
	docURI := protocol.DocumentURI("file:///test.scampi")
	text := `module main
import "std"
import "std/ssh"
import "std/posix"

type Config {
  path: string
  mode: string = "644"
}

let t = ssh.target { name = "web", host = "1.2.3.4", user = "root" }

std.deploy(name = "d", targets = [t]) {
  posix.copy {
    src = posix.source_inline { content = "hello" }
    dest = "/tmp/test"
  }
}
`
	// Not directly testing user types in deploy, but ensuring the
	// general flow works. This is a baseline.
	s.docs.Open(docURI, text, 1)
	items := completionAt(t, s, docURI, 15, 4)
	if len(items) == 0 {
		t.Fatal("expected completions inside deploy body")
	}
}

func TestCompletion_UserType_InForLoop(t *testing.T) {
	s := testServer()
	docURI := protocol.DocumentURI("file:///test.scampi")
	text := `module main

type Box {
  label: string
  color: string
}

func make_boxes() list[Box] {
  return [Box { label = "a", color = "red" }]
}

let boxes = [
  Box { label = "a", color = "red" },
]

func use() string {
  for b in boxes {
    let x = Box {
  }
  return ""
}
`
	s.docs.Open(docURI, text, 1)

	items := completionAt(t, s, docURI, 17, 18)
	if len(items) == 0 {
		t.Fatal("expected field completions for user type in for loop body")
	}
	requireLabels(t, items, "label", "color")
}

func TestCompletion_UserType_WithDefaults(t *testing.T) {
	s := testServer()
	docURI := protocol.DocumentURI("file:///test.scampi")
	text := `module main

type Entry {
  key:   string
  value: string = ""
  ttl:   int = 300
}

let e = Entry {
`
	s.docs.Open(docURI, text, 1)

	items := completionAt(t, s, docURI, 8, 16)
	if len(items) == 0 {
		t.Fatal("expected completions")
	}
	requireLabels(t, items, "key", "value", "ttl")
}

// Struct field dot-access: additional permutations
// -----------------------------------------------------------------------------

func TestCompletion_DotAccess_LetBindingPrefix(t *testing.T) {
	s := testServer()
	docURI := protocol.DocumentURI("file:///test.scampi")
	text := `module main

type Server {
  name: string
  port: int
  tls:  bool = false
}

let srv = Server { name = "web", port = 443 }
let n = srv.na
`
	s.docs.Open(docURI, text, 1)

	// Cursor after "srv.na" — should offer "name" filtered by prefix.
	items := completionAt(t, s, docURI, 9, 14)
	if len(items) == 0 {
		t.Fatal("expected filtered struct field completions")
	}
	requireLabels(t, items, "name")
	rejectLabels(t, items, "port", "tls")
}

func TestCompletion_DotAccess_NoFieldsOnNonStruct(t *testing.T) {
	s := testServer()
	docURI := protocol.DocumentURI("file:///test.scampi")
	text := `module main

let x = "hello"
let y = x.
`
	s.docs.Open(docURI, text, 1)

	// String has no fields — should get UFCS completions or nothing,
	// but NOT crash.
	items := completionAt(t, s, docURI, 3, 10)
	// Just checking it doesn't panic. No field completions expected.
	for _, item := range items {
		if item.Kind == protocol.CompletionItemKindField {
			t.Errorf("strings should not have field completions, got %q", item.Label)
		}
	}
}

func TestCompletion_DotAccess_InsideListLiteral(t *testing.T) {
	s := testServer()
	docURI := protocol.DocumentURI("file:///test.scampi")
	text := `module main

type Cfg {
  host: string
  port: int
}

let c = Cfg { host = "a", port = 1 }
let items = [c.
`
	s.docs.Open(docURI, text, 1)

	items := completionAt(t, s, docURI, 8, 15)
	if len(items) == 0 {
		t.Fatal("expected struct field completions inside list literal")
	}
	requireLabels(t, items, "host", "port")
}

func TestCompletion_DotAccess_FuncParam(t *testing.T) {
	s := testServer()
	docURI := protocol.DocumentURI("file:///test.scampi")
	text := `module main

type Rec {
  a: int
  b: string
}

func process(r: Rec) string {
  let x = r.
  return ""
}
`
	s.docs.Open(docURI, text, 1)

	items := completionAt(t, s, docURI, 8, 12)
	if len(items) == 0 {
		t.Fatal("expected struct field completions for func param")
	}
	requireLabels(t, items, "a", "b")
}

// Kwargs exclusion edge cases
// -----------------------------------------------------------------------------

func TestCompletion_KwargsExclusion_MixedCommasNewlines(t *testing.T) {
	s := testServer()
	docURI := protocol.DocumentURI("file:///test.scampi")
	text := `posix.copy {
  src = posix.source_local { path = "./f" },
  dest = "/etc/foo"
  owner = "root"

}
`
	s.docs.Open(docURI, text, 1)

	// Cursor on empty line 4.
	items := completionAt(t, s, docURI, 4, 2)
	if len(items) == 0 {
		t.Fatal("expected completions")
	}
	rejectLabels(t, items, "src", "dest", "owner")
}

func TestCompletion_KwargsExclusion_CursorAtStart(t *testing.T) {
	s := testServer()
	docURI := protocol.DocumentURI("file:///test.scampi")
	// Cursor right after the opening brace — no fields yet.
	text := "posix.copy {\n  \n}"
	s.docs.Open(docURI, text, 1)

	items := completionAt(t, s, docURI, 1, 2)
	if len(items) == 0 {
		t.Fatal("expected all kwargs when none are present")
	}
	requireLabels(t, items, "src", "dest")
}

func TestCompletion_KwargsExclusion_AllPresent(t *testing.T) {
	s := testServer()
	docURI := protocol.DocumentURI("file:///test.scampi")
	text := `posix.dir {
  path = "/tmp/test"
  state = "present"
  owner = "root"
  group = "root"
  mode = "755"

}
`
	s.docs.Open(docURI, text, 1)

	items := completionAt(t, s, docURI, 6, 2)
	// Most fields are filled — remaining completions should be few.
	rejectLabels(t, items, "path", "state", "owner", "group", "mode")
}

// Real-world scenario: PVE config with for loop
// -----------------------------------------------------------------------------

func TestCompletion_RealWorld_PVEForLoop(t *testing.T) {
	s := testServer()
	docURI := protocol.DocumentURI("file:///test.scampi")
	text := `module main
import "std"
import "std/ssh"
import "std/pve"

type Container {
  id:       int
  hostname: string
  ip:       string
  size:     string
  cores:    int = 1
  memory:   string = "512M"
}

let midgard = ssh.target { name = "midgard", host = "10.10.2.10", user = "root" }
let debian = pve.Template { storage = "local", name = "debian.tar.zst" }

let containers = [
  Container { id = 999, hostname = "a", ip = "10.0.0.1/24", size = "4G", cores = 2, memory = "1G" },
  Container { id = 998, hostname = "b", ip = "10.0.0.2/24", size = "2G" },
]

std.deploy(name = "pve", targets = [midgard]) {
  for c in containers {
    pve.lxc {
      id       = c.id
      node     = "midgard"
      template = debian
      hostname = c.hostname
      memory   = c.memory
      size     = c.size

    }
  }
}
`
	s.docs.Open(docURI, text, 1)

	// Empty line inside pve.lxc — should offer only missing fields.
	items := completionAt(t, s, docURI, 31, 6)
	if len(items) == 0 {
		t.Fatal("expected completions inside pve.lxc")
	}
	rejectLabels(t, items, "id", "node", "template", "hostname", "memory", "size")

	// Dot-access on for-loop variable inside the struct literal.
	text2 := strings.Replace(text, "      memory   = c.memory", "      memory   = c.", 1)
	s.docs.Open(docURI, text2, 2)

	// Find the line with "= c."
	var dotLine uint32
	for i, line := range strings.Split(text2, "\n") {
		if strings.HasSuffix(strings.TrimSpace(line), "= c.") {
			dotLine = uint32(i)
			break
		}
	}
	if dotLine == 0 {
		t.Fatal("could not find c. line")
	}

	dotCol := uint32(strings.Index(strings.Split(text2, "\n")[dotLine], "c.") + 2)
	items2 := completionAt(t, s, docURI, dotLine, dotCol)
	if len(items2) == 0 {
		t.Fatal("expected struct field completions for 'c.' in for loop")
	}
	requireLabels(t, items2, "id", "hostname", "ip", "size", "cores", "memory")
}

// Completion inside deploy body (bare expression position)
// -----------------------------------------------------------------------------

func TestCompletion_DeployBody_TopLevel(t *testing.T) {
	s := testServer()
	docURI := protocol.DocumentURI("file:///test.scampi")
	text := `module main
import "std"
import "std/ssh"
import "std/posix"

let t = ssh.target { name = "web", host = "1.2.3.4", user = "root" }

std.deploy(name = "d", targets = [t]) {
  pos
}
`
	s.docs.Open(docURI, text, 1)

	items := completionAt(t, s, docURI, 8, 5)
	if len(items) == 0 {
		t.Fatal("expected completions inside deploy body")
	}
	requireLabels(t, items, "posix")
}

// Kwarg value completions (constructors)
// -----------------------------------------------------------------------------

func TestCompletion_KwargValue_StdlibConstructor(t *testing.T) {
	s := testServer()
	docURI := protocol.DocumentURI("file:///test.scampi")
	text := "posix.copy {\n  src = \n}"
	s.docs.Open(docURI, text, 1)

	items := completionAt(t, s, docURI, 1, 8)
	if len(items) == 0 {
		t.Fatal("expected constructor completions for src kwarg value")
	}
	// src expects a SourceRef — should offer source_local, source_inline, etc.
	found := false
	for _, item := range items {
		if strings.Contains(item.Label, "source") {
			found = true
			break
		}
	}
	if !found {
		var all []string
		for _, item := range items {
			all = append(all, item.Label)
		}
		t.Errorf("expected source_* constructors, got %v", all)
	}
}
