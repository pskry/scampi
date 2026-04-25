// SPDX-License-Identifier: GPL-3.0-only

package lsp

import (
	"context"
	"strings"
	"testing"

	"go.lsp.dev/protocol"
)

func TestCompletionKwargsExcludesPresent_RealPVE(t *testing.T) {
	s := testServer()
	docURI := protocol.DocumentURI("file:///test.scampi")

	// Exact reproduction of ~/dev/skrynet/pve.scampi with
	// <<CURSOR_HERE>> replaced by an empty indented line.
	const src = `module main

import "std"
import "std/ssh"
import "std/secrets"
import "std/pve"

let age = secrets.from_age(path = "secrets.age.json")

let midgard = ssh.target {
  name = "midgard"
  host = "10.10.2.10"
  user = "hal9000"
}

type Container {
  id:       int
  hostname: string
  ip:       string
  size:     string
  cores:    int = 1
  memory:   string = "512M"
}

// Shared defaults
// -----------------------------------------------------------------------------

let debian = pve.Template { storage = "local", name = "debian-12-standard_12.12-1_amd64.tar.zst" }
let gw = "10.10.2.1"

let ssh_keys = [
  "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIDUBFaQhgMITOPjYtq6SvDhUzDjLWP2se/nMyQRtQCeF hal9000",
]

// Container definitions
// -----------------------------------------------------------------------------

let containers = [
  Container { id = 999, hostname = "scampi-final", ip = "10.10.2.199/24", size = "4G", cores = 2, memory = "1G" },
  Container { id = 998, hostname = "scampi-test2", ip = "10.10.2.198/24", size = "2G" },
]

// Deploy
// -----------------------------------------------------------------------------

std.deploy(name = "pve", targets = [midgard]) {
  for c in containers {
    pve.lxc {
      id              = c.id

      node            = "midgard"
      template        = debian
      hostname        = c.hostname
      cpu             = pve.Cpu { cores = c.cores }
      memory          = c.memory
      size            = c.size
      features        = pve.Features { nesting = true }
      networks        = [pve.LxcNet { bridge = "vmbr0", ip = c.ip, gw = gw }]
      tags            = ["scampi"]
      ssh_public_keys = ssh_keys
    }
  }

  pve.datacenter {
    console     = pve.Console.xtermjs
    keyboard    = "en-us"
    language    = "en"
    mac_prefix  = "be:ef"
    max_workers = 4
    tags        = [
      pve.Tag { name = "ansible",   fg = "#000000", bg = "#73bbbe" },
      pve.Tag { name = "cac",       fg = "#000000", bg = "#81d983" },
      pve.Tag { name = "manual",    fg = "#ffffff", bg = "#6f74e5" },
      pve.Tag { name = "scampi",    fg = "#ffffff", bg = "#e8722a" },
      pve.Tag { name = "snowflake", fg = "#ffffff", bg = "#d1648a" },
    ]
  }
}
`
	s.docs.Open(docURI, src, 1)

	// Find the cursor line (the empty line between id and node).
	cursorLine := uint32(0)
	for i, line := range strings.Split(src, "\n") {
		if strings.TrimSpace(line) == "" && i > 0 {
			prev := strings.Split(src, "\n")[i-1]
			if strings.Contains(prev, "id") && strings.Contains(prev, "c.id") {
				cursorLine = uint32(i)
				break
			}
		}
	}
	if cursorLine == 0 {
		t.Fatal("could not find cursor line")
	}
	t.Logf("cursor at line %d", cursorLine)

	result, err := s.Completion(context.Background(), &protocol.CompletionParams{
		TextDocumentPositionParams: protocol.TextDocumentPositionParams{
			TextDocument: protocol.TextDocumentIdentifier{URI: docURI},
			Position:     protocol.Position{Line: cursorLine, Character: 6},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if result == nil || len(result.Items) == 0 {
		t.Fatal("expected kwarg completions")
	}

	// These fields are already present in the struct literal and
	// must NOT appear in completions.
	present := map[string]bool{
		"id": true, "node": true, "template": true,
		"hostname": true, "cpu": true, "memory": true,
		"size": true, "features": true, "networks": true,
		"tags": true, "ssh_public_keys": true,
	}

	for _, item := range result.Items {
		if present[item.Label] {
			t.Errorf("%q should be excluded (already present)", item.Label)
		}
	}

	t.Logf("got %d completion items:", len(result.Items))
	for _, item := range result.Items {
		t.Logf("  %q kind=%v", item.Label, item.Kind)
	}
}
