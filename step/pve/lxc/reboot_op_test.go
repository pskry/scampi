// SPDX-License-Identifier: GPL-3.0-only

package lxc

import (
	"context"
	"testing"

	"scampi.dev/scampi/capability"
	"scampi.dev/scampi/spec"
	"scampi.dev/scampi/target"
)

type mockTarget struct {
	handler func(cmd string) (target.CommandResult, error)
}

func (m *mockTarget) Capabilities() capability.Capability {
	return capability.PVE | capability.Command
}

func (m *mockTarget) RunCommand(_ context.Context, cmd string) (target.CommandResult, error) {
	return m.handler(cmd)
}

func (m *mockTarget) RunPrivileged(_ context.Context, cmd string) (target.CommandResult, error) {
	return m.handler(cmd)
}

func TestRebootOp_RunsChecksAndDetectsDrift(t *testing.T) {
	cmdr := &mockTarget{handler: func(cmd string) (target.CommandResult, error) {
		if cmd == "pct status 100" {
			return target.CommandResult{Stdout: "status: running\n"}, nil
		}
		return target.CommandResult{ExitCode: 1}, nil
	}}

	op := &rebootLxcOp{
		pveCmd: pveCmd{id: 100},
		checks: []rebootCheck{
			{field: "test", desired: "wanted", probe: func(_ context.Context, _ target.Command, _ int) string {
				return "actual"
			}},
		},
	}
	result, drift, err := op.checkWith(context.Background(), cmdr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != spec.CheckUnsatisfied {
		t.Errorf("got %v, want CheckUnsatisfied", result)
	}
	if len(drift) != 1 || drift[0].Field != "test (reboot)" {
		t.Errorf("got drift %v, want test (reboot)", drift)
	}
}

func TestRebootOp_NoDriftWhenMatched(t *testing.T) {
	cmdr := &mockTarget{handler: func(cmd string) (target.CommandResult, error) {
		if cmd == "pct status 100" {
			return target.CommandResult{Stdout: "status: running\n"}, nil
		}
		return target.CommandResult{ExitCode: 1}, nil
	}}

	op := &rebootLxcOp{
		pveCmd: pveCmd{id: 100},
		checks: []rebootCheck{
			{field: "test", desired: "same", probe: func(_ context.Context, _ target.Command, _ int) string {
				return "same"
			}},
		},
	}
	result, _, err := op.checkWith(context.Background(), cmdr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != spec.CheckSatisfied {
		t.Errorf("got %v, want CheckSatisfied", result)
	}
}

func TestRebootOp_SkippedWhenStopped(t *testing.T) {
	cmdr := &mockTarget{handler: func(cmd string) (target.CommandResult, error) {
		if cmd == "pct status 100" {
			return target.CommandResult{Stdout: "status: stopped\n"}, nil
		}
		return target.CommandResult{ExitCode: 1}, nil
	}}

	op := &rebootLxcOp{
		pveCmd: pveCmd{id: 100},
		checks: []rebootCheck{
			{field: "test", desired: "wanted", probe: func(_ context.Context, _ target.Command, _ int) string {
				return "different"
			}},
		},
	}
	result, _, err := op.checkWith(context.Background(), cmdr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != spec.CheckSatisfied {
		t.Error("got unsatisfied, want CheckSatisfied (no reboot for stopped)")
	}
}

// Op-level reboot check tests
// -----------------------------------------------------------------------------

func TestConfigOp_RebootChecks_Hostname(t *testing.T) {
	op := &configLxcOp{hostname: "pihole"}
	checks := op.RebootChecks()

	var found bool
	for _, c := range checks {
		if c.field == "hostname" {
			found = true
			if c.desired != "pihole" {
				t.Errorf("desired = %q, want pihole", c.desired)
			}
		}
	}
	if !found {
		t.Error("no hostname reboot check")
	}
}

func TestConfigOp_RebootChecks_Features(t *testing.T) {
	op := &configLxcOp{
		features: &LxcFeatures{Nesting: true, Keyctl: true},
	}
	checks := op.RebootChecks()

	var found bool
	for _, c := range checks {
		if c.field == "features" {
			found = true
			if c.desired != "nesting=1,keyctl=1" {
				t.Errorf("desired = %q", c.desired)
			}
		}
	}
	if !found {
		t.Error("no features reboot check")
	}
}

func TestConfigOp_RebootChecks_NoFeaturesWhenNil(t *testing.T) {
	op := &configLxcOp{hostname: "test"}
	for _, c := range op.RebootChecks() {
		if c.field == "features" {
			t.Error("should not have features check when nil")
		}
	}
}

func TestConfigOp_RebootChecks_DNS(t *testing.T) {
	op := &configLxcOp{
		dns: LxcDNS{Nameserver: "1.1.1.1", Searchdomain: "local"},
	}
	checks := op.RebootChecks()

	var ns, sd bool
	for _, c := range checks {
		if c.field == "nameserver" {
			ns = true
			if c.desired != "1.1.1.1" {
				t.Errorf("nameserver desired = %q", c.desired)
			}
		}
		if c.field == "searchdomain" {
			sd = true
			if c.desired != "local" {
				t.Errorf("searchdomain desired = %q", c.desired)
			}
		}
	}
	if !ns {
		t.Error("no nameserver check")
	}
	if !sd {
		t.Error("no searchdomain check")
	}
}

func TestDeviceOp_RebootChecks(t *testing.T) {
	op := &deviceLxcOp{
		devices: []LxcDevice{{Path: "/dev/dri/renderD128", Mode: "0666"}},
	}
	checks := op.RebootChecks()
	if len(checks) != 1 {
		t.Fatalf("got %d checks, want 1", len(checks))
	}
	if checks[0].field != "devices" {
		t.Errorf("field = %q, want devices", checks[0].field)
	}
}

func TestNetworkOp_NoRebootChecks(t *testing.T) {
	op := &networkLxcOp{
		networks: []LxcNet{{Bridge: "vmbr0", IP: "10.0.0.1/24"}},
	}
	if _, ok := any(op).(rebootAware); ok {
		t.Error("networkOp should not be rebootAware")
	}
}
