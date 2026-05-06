// SPDX-License-Identifier: GPL-3.0-only

package pve

import (
	"context"
	"errors"
	"testing"

	"scampi.dev/scampi/capability"
	"scampi.dev/scampi/target"
	"scampi.dev/scampi/target/pkgmgr"
)

// fakeRunner returns a synthesized command result based on a per-cmd
// table. Lookup is by exact match; unknown commands return ErrNotFound.
func fakeRunner(table map[string]target.CommandResult) func(context.Context, string) (target.CommandResult, error) {
	return func(_ context.Context, cmd string) (target.CommandResult, error) {
		if r, ok := table[cmd]; ok {
			return r, nil
		}
		return target.CommandResult{}, errors.New("no fake response")
	}
}

func TestCapabilitiesReturnsStaticCeiling(t *testing.T) {
	tgt := &LXCTarget{vmid: 1000}

	got := tgt.Capabilities()
	want := capability.POSIX |
		capability.Pkg | capability.PkgUpdate | capability.PkgRepo |
		capability.Service | capability.Container

	if got != want {
		t.Errorf("Capabilities() = %v, want %v", got, want)
	}
}

func TestCapabilitiesIgnoresProbeFailure(t *testing.T) {
	tgt := &LXCTarget{vmid: 1000}
	// All backends nil — would have been POSIX-only under probe-derived
	// caps. Static ceiling is independent of detection state.
	if !tgt.Capabilities().HasAll(capability.Pkg | capability.Service | capability.Container) {
		t.Error("expected static ceiling regardless of probe state")
	}
}

func TestEnsureBackendsLxcUnreachable(t *testing.T) {
	tgt := &LXCTarget{vmid: 1000}
	tgt.Runner = fakeRunner(nil) // every probe errors

	err := tgt.ensureBackends(t.Context())
	var unreach LxcUnreachableError
	if !errors.As(err, &unreach) {
		t.Fatalf("expected LxcUnreachableError, got %T: %v", err, err)
	}
	if unreach.VMID != 1000 {
		t.Errorf("VMID = %d, want 1000", unreach.VMID)
	}
}

func TestEnsureBackendsShortCircuitsWhenAlreadyDetected(t *testing.T) {
	calls := 0
	tgt := &LXCTarget{vmid: 1000}
	tgt.Runner = func(_ context.Context, _ string) (target.CommandResult, error) {
		calls++
		return target.CommandResult{}, errors.New("should not be called")
	}
	tgt.PkgBackend = &pkgmgr.Backend{Kind: pkgmgr.Apt}

	if err := tgt.ensureBackends(t.Context()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if calls != 0 {
		t.Errorf("Runner called %d times, expected 0 (already detected)", calls)
	}
}

func TestEnsureBackendsDetectsSystemd(t *testing.T) {
	tgt := &LXCTarget{vmid: 1000}
	// uname errors → skip os-release ReadFile; Platform stays Unknown
	// → PkgBackend nil. systemctl probe via svcmgr.Detect succeeds.
	tgt.Runner = fakeRunner(map[string]target.CommandResult{
		"command -v systemctl":  {ExitCode: 0},
		"command -v rc-service": {ExitCode: 1},
		"command -v launchctl":  {ExitCode: 1},
		"command -v docker":     {ExitCode: 1},
		"command -v podman":     {ExitCode: 1},
		"command -v incus":      {ExitCode: 1},
		"command -v lxc":        {ExitCode: 1},
	})

	if err := tgt.ensureBackends(t.Context()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tgt.SvcBackend == nil {
		t.Error("expected SvcBackend to be populated")
	}
	if tgt.PkgBackend != nil {
		t.Error("expected PkgBackend nil (Platform unknown without os-release)")
	}
}

func TestRequirePkgErrorsWhenBackendMissing(t *testing.T) {
	tgt := &LXCTarget{vmid: 1000}
	// ensureBackends will succeed because SvcBackend gets set, but
	// PkgBackend stays nil — requirePkg must return BackendMissingError.
	tgt.Runner = fakeRunner(map[string]target.CommandResult{
		"command -v systemctl": {ExitCode: 0},
	})

	err := tgt.requirePkg(t.Context())
	var missing BackendMissingError
	if !errors.As(err, &missing) {
		t.Fatalf("expected BackendMissingError, got %T: %v", err, err)
	}
	if missing.Kind != "package manager" {
		t.Errorf("Kind = %q, want %q", missing.Kind, "package manager")
	}
}

func TestRequireSvcErrorsWhenBackendMissing(t *testing.T) {
	tgt := &LXCTarget{vmid: 1000}
	tgt.PkgBackend = &pkgmgr.Backend{Kind: pkgmgr.Apt} // satisfies ensureBackends short-circuit

	err := tgt.requireSvc(t.Context())
	var missing BackendMissingError
	if !errors.As(err, &missing) {
		t.Fatalf("expected BackendMissingError, got %T: %v", err, err)
	}
	if missing.Kind != "service manager" {
		t.Errorf("Kind = %q, want %q", missing.Kind, "service manager")
	}
}

func TestRequireCtrErrorsWhenBackendMissing(t *testing.T) {
	tgt := &LXCTarget{vmid: 1000}
	tgt.PkgBackend = &pkgmgr.Backend{Kind: pkgmgr.Apt}

	err := tgt.requireCtr(t.Context())
	var missing BackendMissingError
	if !errors.As(err, &missing) {
		t.Fatalf("expected BackendMissingError, got %T: %v", err, err)
	}
	if missing.Kind != "container runtime" {
		t.Errorf("Kind = %q, want %q", missing.Kind, "container runtime")
	}
}

// IsInstalled is the canonical PkgManager method override; verifying
// it surfaces the ensureBackends error covers the same wrapper shape
// applied to InstallPkgs/RemovePkgs/UpdateCache/etc.
func TestIsInstalledSurfacesUnreachableError(t *testing.T) {
	tgt := &LXCTarget{vmid: 1000}
	tgt.Runner = fakeRunner(nil)

	_, err := tgt.IsInstalled(t.Context(), "vim")
	var unreach LxcUnreachableError
	if !errors.As(err, &unreach) {
		t.Fatalf("expected LxcUnreachableError, got %T: %v", err, err)
	}
}
