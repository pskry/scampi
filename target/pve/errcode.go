// SPDX-License-Identifier: GPL-3.0-only

package pve

import "scampi.dev/scampi/errs"

const (
	CodePctFailed      errs.Code = "target.pve.PctFailed"
	CodeLxcUnreachable errs.Code = "target.pve.LxcUnreachable"
	CodeBackendMissing errs.Code = "target.pve.BackendMissing"
)
