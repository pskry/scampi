// SPDX-License-Identifier: GPL-3.0-only

package fileops

import (
	"context"
	"fmt"
	"math/rand/v2"
	"strings"

	"scampi.dev/scampi/target"
)

// VerifiedWrite writes content to a temp file, runs verifyCmd against it,
// and only writes to dest if the command exits 0. The temp file is always
// cleaned up. The verifyCmd must contain %s which is replaced with the
// temp file path.
func VerifiedWrite(
	ctx context.Context,
	tgt target.Target,
	dest string,
	content []byte,
	verifyCmd string,
) error {
	fsTgt := target.Must[target.Filesystem]("verify", tgt)
	cmdTgt := target.Must[target.Command]("verify", tgt)

	tmp := tempPath()

	if err := fsTgt.WriteFile(ctx, tmp, content); err != nil {
		return fmt.Errorf("write temp file for verify: %w", err)
	}
	defer func() { _ = fsTgt.Remove(ctx, tmp) }()

	cmd := strings.Replace(verifyCmd, "%s", tmp, 1)
	result, err := cmdTgt.RunCommand(ctx, cmd)
	if err != nil {
		return fmt.Errorf("verify command: %w", err)
	}
	if result.ExitCode != 0 {
		return &VerifyError{
			Cmd:      verifyCmd,
			Dest:     dest,
			ExitCode: result.ExitCode,
			Stderr:   result.Stderr,
		}
	}

	return fsTgt.WriteFile(ctx, dest, content)
}

func tempPath() string {
	return fmt.Sprintf("/tmp/.scampi-%016x", rand.Uint64())
}
