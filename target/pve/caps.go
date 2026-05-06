// SPDX-License-Identifier: GPL-3.0-only

package pve

import (
	"context"
	"strings"
	"time"

	"scampi.dev/scampi/capability"
	"scampi.dev/scampi/target"
	"scampi.dev/scampi/target/ctrmgr"
	"scampi.dev/scampi/target/pkgmgr"
	"scampi.dev/scampi/target/svcmgr"
)

// Capabilities reports the static ceiling of pve.lxc_target rather
// than what Create-time probes managed to detect.
//
// pct exec into a Linux LXC always supports POSIX | Pkg | PkgUpdate |
// PkgRepo | Service | Container — that's the contract of the target
// type, fixed by design. Probe-derived caps were a phase-ordering
// trap: deploy plans run concurrently, so the "configure dc1" target's
// Create races the "create LXC 1000" op in a sibling deploy. When
// Create wins the race, probes return error, all backends stay nil,
// Capabilities() reports POSIX-only, and plan-time aborts:
//
//	missing: Pkg, required: Pkg, provided: POSIX
//
// even though the LXC will be running by the time the configure block
// actually executes. Returning the static ceiling lets the plan check
// pass; backend selection is deferred to first use via ensureBackends.
// See #274.
func (t *LXCTarget) Capabilities() capability.Capability {
	return capability.POSIX |
		capability.Pkg | capability.PkgUpdate | capability.PkgRepo |
		capability.Service | capability.Container
}

// ensureBackends populates PkgBackend / SvcBackend / CtrBackend by
// retrying the in-container probes if Create-time detection couldn't
// reach the LXC. By the time a configure step actually runs, the
// sibling deploy that creates the LXC has finished and probes succeed.
//
// At least one non-nil backend is treated as "detection succeeded";
// further calls are no-ops. If all three are nil we retry — this lets
// the first step that actually touches the LXC drive detection.
func (t *LXCTarget) ensureBackends(ctx context.Context) error {
	t.detectMu.Lock()
	defer t.detectMu.Unlock()
	if t.PkgBackend != nil || t.SvcBackend != nil || t.CtrBackend != nil {
		return nil
	}
	t.detectBackends(ctx)
	if t.PkgBackend == nil && t.SvcBackend == nil && t.CtrBackend == nil {
		return LxcUnreachableError{VMID: t.vmid}
	}
	return nil
}

// detectBackends runs the in-container probes and populates OSInfo +
// PkgBackend / SvcBackend / CtrBackend. Failures (LXC unreachable,
// missing tooling) leave the corresponding fields nil — callers
// decide whether that's fatal. Used by both Create (initial
// best-effort detection) and ensureBackends (lazy retry).
func (t *LXCTarget) detectBackends(ctx context.Context) {
	if r, err := t.Runner(ctx, "uname -s"); err == nil {
		t.OSInfo.Platform = target.ParseKernel(strings.TrimSpace(r.Stdout))
		if osr, err := t.ReadFile(ctx, "/etc/os-release"); err == nil {
			t.OSInfo = target.ResolveLinuxPlatform(osr)
		}
	}
	t.PkgBackend = pkgmgr.Detect(t.OSInfo.Platform)
	detect := func(cmd string) (int, error) {
		r, err := t.Runner(ctx, cmd)
		if err != nil {
			return -1, err
		}
		return r.ExitCode, nil
	}
	t.SvcBackend = svcmgr.Detect(detect)
	t.CtrBackend = ctrmgr.Detect(detect)
}

// requirePkg / requireSvc / requireCtr drive lazy detection and
// produce a clean error if the specific backend the caller needs
// isn't available on this LXC (e.g. invoking posix.pkg on a busybox
// LXC that has neither apt nor apk).
func (t *LXCTarget) requirePkg(ctx context.Context) error {
	if err := t.ensureBackends(ctx); err != nil {
		return err
	}
	if t.PkgBackend == nil {
		return BackendMissingError{VMID: t.vmid, Kind: "package manager"}
	}
	return nil
}

func (t *LXCTarget) requireSvc(ctx context.Context) error {
	if err := t.ensureBackends(ctx); err != nil {
		return err
	}
	if t.SvcBackend == nil {
		return BackendMissingError{VMID: t.vmid, Kind: "service manager"}
	}
	return nil
}

func (t *LXCTarget) requireCtr(ctx context.Context) error {
	if err := t.ensureBackends(ctx); err != nil {
		return err
	}
	if t.CtrBackend == nil {
		return BackendMissingError{VMID: t.vmid, Kind: "container runtime"}
	}
	return nil
}

// PkgManager
// -----------------------------------------------------------------------------

func (t *LXCTarget) IsInstalled(ctx context.Context, pkg string) (bool, error) {
	if err := t.requirePkg(ctx); err != nil {
		return false, err
	}
	return t.Base.IsInstalled(ctx, pkg)
}

func (t *LXCTarget) InstallPkgs(ctx context.Context, pkgs []string) error {
	if err := t.requirePkg(ctx); err != nil {
		return err
	}
	return t.Base.InstallPkgs(ctx, pkgs)
}

func (t *LXCTarget) RemovePkgs(ctx context.Context, pkgs []string) error {
	if err := t.requirePkg(ctx); err != nil {
		return err
	}
	return t.Base.RemovePkgs(ctx, pkgs)
}

// PkgUpdater
// -----------------------------------------------------------------------------

func (t *LXCTarget) UpdateCache(ctx context.Context) error {
	if err := t.requirePkg(ctx); err != nil {
		return err
	}
	return t.Base.UpdateCache(ctx)
}

func (t *LXCTarget) IsUpgradable(ctx context.Context, pkg string) (bool, error) {
	if err := t.requirePkg(ctx); err != nil {
		return false, err
	}
	return t.Base.IsUpgradable(ctx, pkg)
}

func (t *LXCTarget) CacheAge(ctx context.Context) (time.Duration, error) {
	if err := t.requirePkg(ctx); err != nil {
		return 0, err
	}
	return t.Base.CacheAge(ctx)
}

// ServiceManager
// -----------------------------------------------------------------------------

func (t *LXCTarget) IsActive(ctx context.Context, name string) (bool, error) {
	if err := t.requireSvc(ctx); err != nil {
		return false, err
	}
	return t.Base.IsActive(ctx, name)
}

func (t *LXCTarget) IsEnabled(ctx context.Context, name string) (bool, error) {
	if err := t.requireSvc(ctx); err != nil {
		return false, err
	}
	return t.Base.IsEnabled(ctx, name)
}

func (t *LXCTarget) Start(ctx context.Context, name string) error {
	if err := t.requireSvc(ctx); err != nil {
		return err
	}
	return t.Base.Start(ctx, name)
}

func (t *LXCTarget) Stop(ctx context.Context, name string) error {
	if err := t.requireSvc(ctx); err != nil {
		return err
	}
	return t.Base.Stop(ctx, name)
}

func (t *LXCTarget) Enable(ctx context.Context, name string) error {
	if err := t.requireSvc(ctx); err != nil {
		return err
	}
	return t.Base.Enable(ctx, name)
}

func (t *LXCTarget) Disable(ctx context.Context, name string) error {
	if err := t.requireSvc(ctx); err != nil {
		return err
	}
	return t.Base.Disable(ctx, name)
}

func (t *LXCTarget) Restart(ctx context.Context, name string) error {
	if err := t.requireSvc(ctx); err != nil {
		return err
	}
	return t.Base.Restart(ctx, name)
}

func (t *LXCTarget) Reload(ctx context.Context, name string) error {
	if err := t.requireSvc(ctx); err != nil {
		return err
	}
	return t.Base.Reload(ctx, name)
}

func (t *LXCTarget) DaemonReload(ctx context.Context) error {
	if err := t.requireSvc(ctx); err != nil {
		return err
	}
	return t.Base.DaemonReload(ctx)
}

// ContainerManager
// -----------------------------------------------------------------------------

func (t *LXCTarget) InspectContainer(ctx context.Context, name string) (target.ContainerInfo, bool, error) {
	if err := t.requireCtr(ctx); err != nil {
		return target.ContainerInfo{}, false, err
	}
	return t.Base.InspectContainer(ctx, name)
}

func (t *LXCTarget) CreateContainer(ctx context.Context, opts target.ContainerInfo) error {
	if err := t.requireCtr(ctx); err != nil {
		return err
	}
	return t.Base.CreateContainer(ctx, opts)
}

func (t *LXCTarget) StartContainer(ctx context.Context, name string) error {
	if err := t.requireCtr(ctx); err != nil {
		return err
	}
	return t.Base.StartContainer(ctx, name)
}

func (t *LXCTarget) StopContainer(ctx context.Context, name string) error {
	if err := t.requireCtr(ctx); err != nil {
		return err
	}
	return t.Base.StopContainer(ctx, name)
}

func (t *LXCTarget) RemoveContainer(ctx context.Context, name string) error {
	if err := t.requireCtr(ctx); err != nil {
		return err
	}
	return t.Base.RemoveContainer(ctx, name)
}
