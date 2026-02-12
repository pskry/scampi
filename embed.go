// SPDX-License-Identifier: GPL-3.0-only

package doit

import "embed"

//go:embed cue/**
var EmbeddedSchemaModule embed.FS
