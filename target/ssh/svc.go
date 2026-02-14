// SPDX-License-Identifier: GPL-3.0-only

package ssh

import (
	"context"
	"fmt"

	"godoit.dev/doit/target"
)

func (t *SSHTarget) IsActive(ctx context.Context, name string) (bool, error) {
	cmd := fmt.Sprintf(t.svcBackend.IsActive, shellQuote(name))
	result, err := t.RunCommand(ctx, cmd)
	if err != nil {
		return false, err
	}
	return result.ExitCode == 0, nil
}

func (t *SSHTarget) IsEnabled(ctx context.Context, name string) (bool, error) {
	cmd := fmt.Sprintf(t.svcBackend.IsEnabled, shellQuote(name))
	result, err := t.RunCommand(ctx, cmd)
	if err != nil {
		return false, err
	}
	return result.ExitCode == 0, nil
}

func (t *SSHTarget) Start(ctx context.Context, name string) error {
	return t.runSvcCommand(ctx, t.svcBackend.Start, name, "start")
}

func (t *SSHTarget) Stop(ctx context.Context, name string) error {
	return t.runSvcCommand(ctx, t.svcBackend.Stop, name, "stop")
}

func (t *SSHTarget) Enable(ctx context.Context, name string) error {
	return t.runSvcCommand(ctx, t.svcBackend.Enable, name, "enable")
}

func (t *SSHTarget) Disable(ctx context.Context, name string) error {
	return t.runSvcCommand(ctx, t.svcBackend.Disable, name, "disable")
}

func (t *SSHTarget) DaemonReload(ctx context.Context) error {
	if t.svcBackend.DaemonReload == "" {
		return nil
	}
	if t.svcBackend.NeedsRoot && !t.isRoot && t.escalate == "" {
		return target.NoEscalationError{Op: t.svcBackend.Name + " daemon-reload"}
	}
	cmd := t.svcBackend.DaemonReload
	if t.svcBackend.NeedsRoot && t.escalate != "" {
		cmd = t.escalate + " " + cmd
	}
	result, err := t.RunCommand(ctx, cmd)
	if err != nil {
		return err
	}
	if result.ExitCode != 0 {
		return ServiceCommandError{
			Op:       "daemon-reload",
			Stderr:   result.Stderr,
			ExitCode: result.ExitCode,
		}
	}
	return nil
}

func (t *SSHTarget) runSvcCommand(ctx context.Context, tmpl, name, op string) error {
	if t.svcBackend.NeedsRoot && !t.isRoot && t.escalate == "" {
		return target.NoEscalationError{Op: t.svcBackend.Name + " " + op}
	}
	cmd := fmt.Sprintf(tmpl, shellQuote(name))
	if t.svcBackend.NeedsRoot && t.escalate != "" {
		cmd = t.escalate + " " + cmd
	}
	result, err := t.RunCommand(ctx, cmd)
	if err != nil {
		return err
	}
	if result.ExitCode != 0 {
		return ServiceCommandError{
			Op:       op,
			Name:     name,
			Stderr:   result.Stderr,
			ExitCode: result.ExitCode,
		}
	}
	return nil
}

// ServiceCommandError is returned when a service command fails.
type ServiceCommandError struct {
	Op       string
	Name     string
	Stderr   string
	ExitCode int
}

func (e ServiceCommandError) Error() string {
	return fmt.Sprintf("service %s %s failed (exit %d): %s", e.Op, e.Name, e.ExitCode, e.Stderr)
}
