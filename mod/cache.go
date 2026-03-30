// SPDX-License-Identifier: GPL-3.0-only

package mod

import (
	"os"
	"path/filepath"
)

func defaultCacheDir() string {
	if xdg := os.Getenv("XDG_CACHE_HOME"); xdg != "" {
		return filepath.Join(xdg, "scampi", "mod")
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(".cache", "scampi", "mod")
	}
	return filepath.Join(home, ".cache", "scampi", "mod")
}
