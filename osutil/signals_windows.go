//go:build windows

package osutil

import "os"

var MainContextSignals = []os.Signal{
	os.Interrupt, // Ctrl+C
}
