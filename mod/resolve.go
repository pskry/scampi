// SPDX-License-Identifier: GPL-3.0-only

package mod

import (
	"os"
	"path/filepath"
	"strings"
)

// Resolve maps a module load path to an absolute .star file path in the cache.
//
// For bare module loads (e.g. codeberg.org/user/repo), the entry point is
// resolved by trying _index.star then <last-segment>.star. If both exist,
// _index.star takes precedence.
//
// For subpath loads (e.g. codeberg.org/user/repo/internal/helpers), the
// subpath is resolved by trying <subpath>.star then <subpath>/_index.star.
func Resolve(m *Module, loadPath string, cacheDir string) (string, error) {
	dep, subPath := splitModulePath(m, loadPath)
	if dep == nil {
		return "", &ModuleNotFoundError{LoadPath: loadPath}
	}

	modDir := filepath.Join(cacheDir, dep.Path+"@"+dep.Version)
	if _, err := os.Stat(modDir); os.IsNotExist(err) {
		depCopy := dep
		return "", &ModuleNotCachedError{
			ModPath: dep.Path,
			Version: dep.Version,
			Source:  m.DepSpan(depCopy),
		}
	}

	candidates, err := resolveCandidates(modDir, dep, subPath)
	if err != nil {
		return "", err
	}

	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}
	}

	tried := make([]string, len(candidates))
	for i, c := range candidates {
		tried[i] = strings.TrimPrefix(c, modDir+string(filepath.Separator))
	}
	return "", &ModuleNoEntryPointError{
		ModPath: dep.Path,
		Tried:   tried,
	}
}

// resolveCandidates returns the ordered list of candidate .star paths to try.
func resolveCandidates(modDir string, dep *Dependency, subPath string) ([]string, error) {
	if subPath == "" {
		last := lastSegment(dep.Path)
		return []string{
			filepath.Join(modDir, "_index.star"),
			filepath.Join(modDir, last+".star"),
		}, nil
	}

	subNative := filepath.FromSlash(subPath)
	return []string{
		filepath.Join(modDir, subNative+".star"),
		filepath.Join(modDir, subNative, "_index.star"),
	}, nil
}

// splitModulePath finds the longest require-table prefix of loadPath and
// returns the matching Dependency and any remaining subpath.
func splitModulePath(m *Module, loadPath string) (*Dependency, string) {
	var best *Dependency
	for i := range m.Require {
		dep := &m.Require[i]
		if loadPath == dep.Path {
			return dep, ""
		}
		if strings.HasPrefix(loadPath, dep.Path+"/") {
			if best == nil || len(dep.Path) > len(best.Path) {
				best = dep
			}
		}
	}
	if best != nil {
		sub := strings.TrimPrefix(loadPath, best.Path+"/")
		return best, sub
	}
	return nil, ""
}

// lastSegment returns the last path segment of a slash-separated path.
func lastSegment(p string) string {
	if i := strings.LastIndex(p, "/"); i >= 0 {
		return p[i+1:]
	}
	return p
}

// DefaultCacheDir returns the default module cache directory.
// Uses $XDG_CACHE_HOME/scampi/mod if set, else ~/.cache/scampi/mod.
func DefaultCacheDir() string {
	if xdg := os.Getenv("XDG_CACHE_HOME"); xdg != "" {
		return filepath.Join(xdg, "scampi", "mod")
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(".cache", "scampi", "mod")
	}
	return filepath.Join(home, ".cache", "scampi", "mod")
}
