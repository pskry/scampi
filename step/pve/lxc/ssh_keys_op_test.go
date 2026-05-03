// SPDX-License-Identifier: GPL-3.0-only

package lxc

import (
	"context"
	"strings"
	"testing"

	"scampi.dev/scampi/target"
)

// TestSSHKeyDrift_CheckIsReadOnly is the regression for #242: the
// Check phase must not mutate target state. The pre-fix implementation
// pulled a temp file (`pct pull` writes), cat-ed it, then rm-ed it —
// three mutations on the PVE host per Check. The fix uses a single
// `pct exec … cat …`, which writes nothing.
//
// Drift-detection scenarios live in TestSSHKeyDrift (pct_test.go); this
// test asserts only the read-only contract.
func TestSSHKeyDrift_CheckIsReadOnly(t *testing.T) {
	var commands []string
	cmdr := &mockTarget{handler: func(cmd string) (target.CommandResult, error) {
		commands = append(commands, cmd)
		if strings.HasPrefix(cmd, "pct exec 100 -- cat /root/.ssh/authorized_keys") {
			return target.CommandResult{
				Stdout: "# --- BEGIN PVE ---\nssh-rsa AAAA test1\n# --- END PVE ---\n",
			}, nil
		}
		return target.CommandResult{ExitCode: 1}, nil
	}}

	op := &sshKeysLxcOp{
		pveCmd:        pveCmd{id: 100},
		sshPublicKeys: []string{"ssh-rsa AAAA test1"},
	}
	_ = op.sshKeyDrift(context.Background(), cmdr)

	for _, cmd := range commands {
		switch {
		case strings.Contains(cmd, "pct pull"):
			t.Errorf("Check issued pct pull (mutating): %q", cmd)
		case strings.Contains(cmd, "/tmp/.scampi-ssh-read"):
			t.Errorf("Check referenced temp file path: %q", cmd)
		case strings.HasPrefix(cmd, "rm "):
			t.Errorf("Check issued rm: %q", cmd)
		}
	}
	if len(commands) != 1 {
		t.Errorf("expected 1 command (single pct exec cat), got %d: %v", len(commands), commands)
	}
}
