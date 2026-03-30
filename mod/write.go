// SPDX-License-Identifier: GPL-3.0-only

package mod

import (
	"context"
	"fmt"
	"strings"

	"scampi.dev/scampi/source"
)

// writeModFile serialises module and deps to path in canonical scampi.mod format.
func writeModFile(ctx context.Context, src source.Source, path, module string, deps []Dependency) error {
	var sb strings.Builder
	sb.WriteString("module ")
	sb.WriteString(module)
	sb.WriteString("\n")

	if len(deps) > 0 {
		sb.WriteString("\nrequire (\n")
		for _, dep := range deps {
			sb.WriteString("\t")
			sb.WriteString(dep.Path)
			sb.WriteString(" ")
			sb.WriteString(dep.Version)
			sb.WriteString("\n")
		}
		sb.WriteString(")\n")
	}

	if err := src.WriteFile(ctx, path, []byte(sb.String())); err != nil {
		return &WriteError{
			Detail: fmt.Sprintf("could not write scampi.mod: %v", err),
			Hint:   "check file permissions",
		}
	}
	return nil
}
