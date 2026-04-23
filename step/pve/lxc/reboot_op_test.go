// SPDX-License-Identifier: GPL-3.0-only

package lxc

import (
	"context"
	"testing"

	"scampi.dev/scampi/capability"
	"scampi.dev/scampi/spec"
	"scampi.dev/scampi/target"
)

// mockTarget implements target.Target + target.Command for testing ops.
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

func TestRebootOpCheck_NoRebootNeeded(t *testing.T) {
	cmdr := &mockTarget{handler: func(cmd string) (target.CommandResult, error) {
		switch cmd {
		case "pct status 100":
			return target.CommandResult{Stdout: "status: running\n"}, nil
		case "pct exec 100 -- hostname":
			return target.CommandResult{Stdout: "pihole\n"}, nil
		default:
			return target.CommandResult{ExitCode: 1}, nil
		}
	}}

	op := &rebootLxcOp{pveCmd: pveCmd{id: 100}, hostname: "pihole"}
	result, drift, err := op.checkWith(context.Background(), cmdr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != spec.CheckSatisfied {
		t.Errorf("got %v, want CheckSatisfied", result)
	}
	if len(drift) != 0 {
		t.Errorf("got %d drift entries, want 0", len(drift))
	}
}

func TestRebootOpCheck_HostnameMismatch(t *testing.T) {
	cmdr := &mockTarget{handler: func(cmd string) (target.CommandResult, error) {
		switch cmd {
		case "pct status 100":
			return target.CommandResult{Stdout: "status: running\n"}, nil
		case "pct exec 100 -- hostname":
			return target.CommandResult{Stdout: "old-name\n"}, nil
		default:
			return target.CommandResult{ExitCode: 1}, nil
		}
	}}

	op := &rebootLxcOp{pveCmd: pveCmd{id: 100}, hostname: "pihole"}
	result, drift, err := op.checkWith(context.Background(), cmdr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != spec.CheckUnsatisfied {
		t.Errorf("got %v, want CheckUnsatisfied", result)
	}
	if len(drift) != 1 || drift[0].Field != "hostname (reboot)" {
		t.Errorf("got drift %v, want hostname (reboot)", drift)
	}
}

func TestRebootOpCheck_ContainerStopped(t *testing.T) {
	cmdr := &mockTarget{handler: func(cmd string) (target.CommandResult, error) {
		switch cmd {
		case "pct status 100":
			return target.CommandResult{Stdout: "status: stopped\n"}, nil
		default:
			return target.CommandResult{ExitCode: 1}, nil
		}
	}}

	op := &rebootLxcOp{pveCmd: pveCmd{id: 100}, hostname: "pihole"}
	result, _, err := op.checkWith(context.Background(), cmdr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != spec.CheckSatisfied {
		t.Errorf("got %v, want CheckSatisfied (no reboot for stopped container)", result)
	}
}

func TestRebootOpCheck_FeaturesDrift(t *testing.T) {
	cmdr := &mockTarget{handler: func(cmd string) (target.CommandResult, error) {
		switch cmd {
		case "pct status 100":
			return target.CommandResult{Stdout: "status: running\n"}, nil
		case "pct exec 100 -- hostname":
			return target.CommandResult{Stdout: "pihole\n"}, nil
		case "pct config 100":
			// Current config has nesting=1 but desired also wants keyctl.
			return target.CommandResult{
				Stdout: "hostname: pihole\nfeatures: nesting=1\n",
			}, nil
		default:
			return target.CommandResult{ExitCode: 1}, nil
		}
	}}

	op := &rebootLxcOp{
		pveCmd:   pveCmd{id: 100},
		hostname: "pihole",
		features: &LxcFeatures{Nesting: true, Keyctl: true},
	}
	result, drift, err := op.checkWith(context.Background(), cmdr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != spec.CheckUnsatisfied {
		t.Errorf("got %v, want CheckUnsatisfied", result)
	}
	if len(drift) != 1 || drift[0].Field != "features (reboot)" {
		t.Errorf("got drift %v, want features (reboot)", drift)
	}
}

func TestRebootOpCheck_FeaturesMatch(t *testing.T) {
	cmdr := &mockTarget{handler: func(cmd string) (target.CommandResult, error) {
		switch cmd {
		case "pct status 100":
			return target.CommandResult{Stdout: "status: running\n"}, nil
		case "pct exec 100 -- hostname":
			return target.CommandResult{Stdout: "pihole\n"}, nil
		case "pct config 100":
			return target.CommandResult{
				Stdout: "hostname: pihole\nfeatures: nesting=1,keyctl=1\n",
			}, nil
		default:
			return target.CommandResult{ExitCode: 1}, nil
		}
	}}

	op := &rebootLxcOp{
		pveCmd:   pveCmd{id: 100},
		hostname: "pihole",
		features: &LxcFeatures{Nesting: true, Keyctl: true},
	}
	result, drift, err := op.checkWith(context.Background(), cmdr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != spec.CheckSatisfied {
		t.Errorf("got %v, want CheckSatisfied", result)
	}
	if len(drift) != 0 {
		t.Errorf("got %d drift entries, want 0", len(drift))
	}
}

func TestRebootOpCheck_DeviceAdded(t *testing.T) {
	cmdr := &mockTarget{handler: func(cmd string) (target.CommandResult, error) {
		switch cmd {
		case "pct status 100":
			return target.CommandResult{Stdout: "status: running\n"}, nil
		case "pct exec 100 -- hostname":
			return target.CommandResult{Stdout: "pihole\n"}, nil
		case "pct config 100":
			return target.CommandResult{Stdout: "hostname: pihole\n"}, nil
		default:
			return target.CommandResult{ExitCode: 1}, nil
		}
	}}

	op := &rebootLxcOp{
		pveCmd:   pveCmd{id: 100},
		hostname: "pihole",
		devices:  []LxcDevice{{Path: "/dev/dri/renderD128", Mode: "0666"}},
	}
	result, drift, err := op.checkWith(context.Background(), cmdr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != spec.CheckUnsatisfied {
		t.Errorf("got %v, want CheckUnsatisfied", result)
	}
	if len(drift) != 1 || drift[0].Field != "devices (reboot)" {
		t.Errorf("got drift %v, want devices (reboot)", drift)
	}
}

func TestRebootOpCheck_DeviceRemoved(t *testing.T) {
	cmdr := &mockTarget{handler: func(cmd string) (target.CommandResult, error) {
		switch cmd {
		case "pct status 100":
			return target.CommandResult{Stdout: "status: running\n"}, nil
		case "pct exec 100 -- hostname":
			return target.CommandResult{Stdout: "pihole\n"}, nil
		case "pct config 100":
			return target.CommandResult{
				Stdout: "hostname: pihole\n" +
					"dev0: /dev/dri/renderD128,mode=0666\n",
			}, nil
		default:
			return target.CommandResult{ExitCode: 1}, nil
		}
	}}

	op := &rebootLxcOp{
		pveCmd:   pveCmd{id: 100},
		hostname: "pihole",
	}
	result, drift, err := op.checkWith(context.Background(), cmdr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != spec.CheckUnsatisfied {
		t.Errorf("got %v, want CheckUnsatisfied", result)
	}
	if len(drift) != 1 || drift[0].Field != "devices (reboot)" {
		t.Errorf("got drift %v, want devices (reboot)", drift)
	}
}

func TestRebootOpCheck_DevicesConverged(t *testing.T) {
	cmdr := &mockTarget{handler: func(cmd string) (target.CommandResult, error) {
		switch cmd {
		case "pct status 100":
			return target.CommandResult{Stdout: "status: running\n"}, nil
		case "pct exec 100 -- hostname":
			return target.CommandResult{Stdout: "pihole\n"}, nil
		case "pct config 100":
			return target.CommandResult{
				Stdout: "hostname: pihole\n" +
					"dev0: /dev/dri/renderD128,mode=0666\n",
			}, nil
		case "pct exec 100 -- test -e '/dev/dri/renderD128'":
			return target.CommandResult{ExitCode: 0}, nil
		default:
			return target.CommandResult{ExitCode: 1}, nil
		}
	}}

	op := &rebootLxcOp{
		pveCmd:   pveCmd{id: 100},
		hostname: "pihole",
		devices:  []LxcDevice{{Path: "/dev/dri/renderD128", Mode: "0666"}},
	}
	result, drift, err := op.checkWith(context.Background(), cmdr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != spec.CheckSatisfied {
		t.Errorf("got %v, want CheckSatisfied", result)
	}
	if len(drift) != 0 {
		t.Errorf("got %d drift entries, want 0", len(drift))
	}
}

func TestRebootOpCheck_DeviceInterrupted(t *testing.T) {
	// Config matches desired but device isn't inside container
	// (interrupted run: cfgOp wrote config, reboot didn't fire).
	cmdr := &mockTarget{handler: func(cmd string) (target.CommandResult, error) {
		switch cmd {
		case "pct status 100":
			return target.CommandResult{Stdout: "status: running\n"}, nil
		case "pct exec 100 -- hostname":
			return target.CommandResult{Stdout: "pihole\n"}, nil
		case "pct config 100":
			return target.CommandResult{
				Stdout: "hostname: pihole\n" +
					"dev0: /dev/dri/renderD128,mode=0666\n",
			}, nil
		case "pct exec 100 -- test -e '/dev/dri/renderD128'":
			return target.CommandResult{ExitCode: 1}, nil
		default:
			return target.CommandResult{ExitCode: 1}, nil
		}
	}}

	op := &rebootLxcOp{
		pveCmd:   pveCmd{id: 100},
		hostname: "pihole",
		devices:  []LxcDevice{{Path: "/dev/dri/renderD128", Mode: "0666"}},
	}
	result, drift, err := op.checkWith(context.Background(), cmdr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != spec.CheckUnsatisfied {
		t.Errorf("got %v, want CheckUnsatisfied", result)
	}
	if len(drift) != 1 || drift[0].Field != "devices (reboot)" {
		t.Errorf("got drift %v, want devices (reboot)", drift)
	}
}

func TestRebootOpCheck_StaleDevice(t *testing.T) {
	// Config and desired both empty, but /dev/dri exists inside
	// container from a previous config (interrupted removal).
	cmdr := &mockTarget{handler: func(cmd string) (target.CommandResult, error) {
		switch cmd {
		case "pct status 100":
			return target.CommandResult{Stdout: "status: running\n"}, nil
		case "pct exec 100 -- hostname":
			return target.CommandResult{Stdout: "pihole\n"}, nil
		case "pct config 100":
			return target.CommandResult{Stdout: "hostname: pihole\n"}, nil
		case "pct exec 100 -- test -d /dev/dri":
			return target.CommandResult{ExitCode: 0}, nil
		default:
			return target.CommandResult{ExitCode: 1}, nil
		}
	}}

	op := &rebootLxcOp{
		pveCmd:   pveCmd{id: 100},
		hostname: "pihole",
	}
	result, drift, err := op.checkWith(context.Background(), cmdr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != spec.CheckUnsatisfied {
		t.Errorf("got %v, want CheckUnsatisfied", result)
	}
	if len(drift) != 1 || drift[0].Field != "devices (reboot)" {
		t.Errorf("got drift %v, want devices (reboot)", drift)
	}
}

func TestRebootOpCheck_NoDevicesClean(t *testing.T) {
	// No devices desired, no devices in config, no /dev/dri in
	// container — fully clean, no reboot.
	cmdr := &mockTarget{handler: func(cmd string) (target.CommandResult, error) {
		switch cmd {
		case "pct status 100":
			return target.CommandResult{Stdout: "status: running\n"}, nil
		case "pct exec 100 -- hostname":
			return target.CommandResult{Stdout: "pihole\n"}, nil
		case "pct config 100":
			return target.CommandResult{Stdout: "hostname: pihole\n"}, nil
		case "pct exec 100 -- test -d /dev/dri":
			return target.CommandResult{ExitCode: 1}, nil
		default:
			return target.CommandResult{ExitCode: 1}, nil
		}
	}}

	op := &rebootLxcOp{
		pveCmd:   pveCmd{id: 100},
		hostname: "pihole",
	}
	result, drift, err := op.checkWith(context.Background(), cmdr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != spec.CheckSatisfied {
		t.Errorf("got %v, want CheckSatisfied", result)
	}
	if len(drift) != 0 {
		t.Errorf("got %d drift entries, want 0", len(drift))
	}
}
