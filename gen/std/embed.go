// SPDX-License-Identifier: GPL-3.0-only

package std

import "embed"

// FS contains the generated scampi-lang stub files for the standard
// library. Embedded at build time, extracted to the global module
// cache on first use.
//
//go:embed *.scampi */*.scampi
var FS embed.FS
