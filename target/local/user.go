// SPDX-License-Identifier: GPL-3.0-only

package local

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"scampi.dev/scampi/target"
)

// UserManager
// -----------------------------------------------------------------------------

func (t POSIXTarget) UserExists(ctx context.Context, name string) (bool, error) {
	result, err := t.RunCommand(ctx, "getent passwd "+target.ShellQuote(name))
	if err != nil {
		return false, err
	}
	return result.ExitCode == 0, nil
}

func (t POSIXTarget) GetUser(ctx context.Context, name string) (target.UserInfo, error) {
	result, err := t.RunCommand(ctx, "getent passwd "+target.ShellQuote(name))
	if err != nil {
		return target.UserInfo{}, err
	}
	if result.ExitCode != 0 {
		return target.UserInfo{}, target.ErrUnknownUser
	}
	info, err := parsePasswdLine(strings.TrimSpace(result.Stdout))
	if err != nil {
		return target.UserInfo{}, err
	}

	// Fetch supplementary groups
	grResult, err := t.RunCommand(ctx, "id -Gn "+target.ShellQuote(name))
	if err == nil && grResult.ExitCode == 0 {
		allGroups := strings.Fields(strings.TrimSpace(grResult.Stdout))
		// Filter out primary group
		pgResult, _ := t.RunCommand(ctx, "id -gn "+target.ShellQuote(name))
		primaryGroup := strings.TrimSpace(pgResult.Stdout)
		var supplementary []string
		for _, g := range allGroups {
			if g != primaryGroup {
				supplementary = append(supplementary, g)
			}
		}
		info.Groups = supplementary
	}

	return info, nil
}

func (t POSIXTarget) CreateUser(ctx context.Context, info target.UserInfo) error {
	if !t.isRoot && t.escalate == "" {
		return target.NoEscalationError{Op: "useradd"}
	}

	cmd := "useradd"
	if info.Shell != "" {
		cmd += " -s " + target.ShellQuote(info.Shell)
	}
	if info.Home != "" {
		cmd += " -m -d " + target.ShellQuote(info.Home)
	}
	if info.System {
		cmd += " -r"
	}
	if info.Password != "" {
		cmd += " -p " + target.ShellQuote(info.Password)
	}
	if len(info.Groups) > 0 {
		cmd += " -G " + target.ShellQuote(strings.Join(info.Groups, ","))
	}
	cmd += " " + target.ShellQuote(info.Name)

	if t.escalate != "" {
		cmd = t.escalate + " " + cmd
	}

	result, err := t.RunCommand(ctx, cmd)
	if err != nil {
		return err
	}
	if result.ExitCode != 0 {
		return fmt.Errorf("useradd %s failed (exit %d): %s", info.Name, result.ExitCode, result.Stderr)
	}
	return nil
}

func (t POSIXTarget) ModifyUser(ctx context.Context, info target.UserInfo) error {
	if !t.isRoot && t.escalate == "" {
		return target.NoEscalationError{Op: "usermod"}
	}

	cmd := "usermod"
	if info.Shell != "" {
		cmd += " -s " + target.ShellQuote(info.Shell)
	}
	if info.Home != "" {
		cmd += " -d " + target.ShellQuote(info.Home)
	}
	if info.Password != "" {
		cmd += " -p " + target.ShellQuote(info.Password)
	}
	if info.Groups != nil {
		cmd += " -G " + target.ShellQuote(strings.Join(info.Groups, ","))
	}
	cmd += " " + target.ShellQuote(info.Name)

	if t.escalate != "" {
		cmd = t.escalate + " " + cmd
	}

	result, err := t.RunCommand(ctx, cmd)
	if err != nil {
		return err
	}
	if result.ExitCode != 0 {
		return fmt.Errorf("usermod %s failed (exit %d): %s", info.Name, result.ExitCode, result.Stderr)
	}
	return nil
}

func (t POSIXTarget) DeleteUser(ctx context.Context, name string) error {
	if !t.isRoot && t.escalate == "" {
		return target.NoEscalationError{Op: "userdel"}
	}

	cmd := "userdel " + target.ShellQuote(name)
	if t.escalate != "" {
		cmd = t.escalate + " " + cmd
	}

	result, err := t.RunCommand(ctx, cmd)
	if err != nil {
		return err
	}
	if result.ExitCode != 0 {
		return fmt.Errorf("userdel %s failed (exit %d): %s", name, result.ExitCode, result.Stderr)
	}
	return nil
}

// GroupManager
// -----------------------------------------------------------------------------

func (t POSIXTarget) GroupExists(ctx context.Context, name string) (bool, error) {
	result, err := t.RunCommand(ctx, "getent group "+target.ShellQuote(name))
	if err != nil {
		return false, err
	}
	return result.ExitCode == 0, nil
}

func (t POSIXTarget) GetGroup(ctx context.Context, name string) (target.GroupInfo, error) {
	result, err := t.RunCommand(ctx, "getent group "+target.ShellQuote(name))
	if err != nil {
		return target.GroupInfo{}, err
	}
	if result.ExitCode != 0 {
		return target.GroupInfo{}, target.ErrUnknownGroup
	}
	return parseGroupLine(strings.TrimSpace(result.Stdout))
}

func (t POSIXTarget) CreateGroup(ctx context.Context, info target.GroupInfo) error {
	if !t.isRoot && t.escalate == "" {
		return target.NoEscalationError{Op: "groupadd"}
	}

	cmd := "groupadd"
	if info.GID != 0 {
		cmd += " -g " + strconv.Itoa(info.GID)
	}
	if info.System {
		cmd += " -r"
	}
	cmd += " " + target.ShellQuote(info.Name)

	if t.escalate != "" {
		cmd = t.escalate + " " + cmd
	}

	result, err := t.RunCommand(ctx, cmd)
	if err != nil {
		return err
	}
	if result.ExitCode != 0 {
		return fmt.Errorf("groupadd %s failed (exit %d): %s", info.Name, result.ExitCode, result.Stderr)
	}
	return nil
}

func (t POSIXTarget) DeleteGroup(ctx context.Context, name string) error {
	if !t.isRoot && t.escalate == "" {
		return target.NoEscalationError{Op: "groupdel"}
	}

	cmd := "groupdel " + target.ShellQuote(name)
	if t.escalate != "" {
		cmd = t.escalate + " " + cmd
	}

	result, err := t.RunCommand(ctx, cmd)
	if err != nil {
		return err
	}
	if result.ExitCode != 0 {
		return fmt.Errorf("groupdel %s failed (exit %d): %s", name, result.ExitCode, result.Stderr)
	}
	return nil
}

// Helpers
// -----------------------------------------------------------------------------

// parsePasswdLine parses a getent passwd line: name:x:uid:gid:gecos:home:shell
func parsePasswdLine(line string) (target.UserInfo, error) {
	parts := strings.Split(line, ":")
	if len(parts) < 7 {
		return target.UserInfo{}, fmt.Errorf("unexpected passwd format: %q", line)
	}
	uid, _ := strconv.Atoi(parts[2])
	gid, _ := strconv.Atoi(parts[3])
	return target.UserInfo{
		Name:  parts[0],
		UID:   uid,
		GID:   gid,
		Home:  parts[5],
		Shell: parts[6],
	}, nil
}

// parseGroupLine parses a getent group line: name:x:gid:member1,member2
func parseGroupLine(line string) (target.GroupInfo, error) {
	parts := strings.Split(line, ":")
	if len(parts) < 4 {
		return target.GroupInfo{}, fmt.Errorf("unexpected group format: %q", line)
	}
	gid, _ := strconv.Atoi(parts[2])
	var members []string
	if parts[3] != "" {
		members = strings.Split(parts[3], ",")
	}
	return target.GroupInfo{
		Name:    parts[0],
		GID:     gid,
		Members: members,
	}, nil
}
