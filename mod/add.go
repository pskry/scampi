// SPDX-License-Identifier: GPL-3.0-only

package mod

import (
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
)

// Add adds or updates a dependency in the scampi.mod and scampi.sum files at dir.
// If version is empty, the latest stable semver tag is resolved from the remote.
// Returns the resolved version.
func Add(modPath, version, dir, cacheDir string) (string, error) {
	if version == "" {
		resolved, err := resolveLatestStable(modPath)
		if err != nil {
			return "", err
		}
		version = resolved
	}

	dep := Dependency{Path: modPath, Version: version}

	if err := Fetch(dep, cacheDir); err != nil {
		return "", err
	}

	dest := filepath.Join(cacheDir, dep.Path+"@"+dep.Version)

	if err := ValidateEntryPoint(dep, dest); err != nil {
		_ = os.RemoveAll(dest)
		return "", err
	}

	hash, err := ComputeHash(dest)
	if err != nil {
		return "", err
	}

	modFile := filepath.Join(dir, "scampi.mod")
	data, err := os.ReadFile(modFile)
	if err != nil {
		return "", &AddError{
			Detail: "could not read scampi.mod: " + err.Error(),
			Hint:   "run: scampi mod init",
		}
	}

	m, err := Parse(modFile, data)
	if err != nil {
		return "", err
	}

	deps := make([]Dependency, 0, len(m.Require)+1)
	added := false
	for _, d := range m.Require {
		if d.Path == modPath {
			deps = append(deps, Dependency{Path: modPath, Version: version})
			added = true
		} else {
			deps = append(deps, Dependency{Path: d.Path, Version: d.Version})
		}
	}
	if !added {
		deps = append(deps, Dependency{Path: modPath, Version: version})
	}

	slices.SortFunc(deps, func(a, b Dependency) int {
		return strings.Compare(a.Path, b.Path)
	})

	if err := writeModFile(modFile, m.Module, deps); err != nil {
		return "", err
	}

	sumFile := filepath.Join(dir, "scampi.sum")
	sums, err := ReadSum(sumFile)
	if err != nil {
		return "", err
	}

	key := modPath + " " + version
	sums[key] = hash

	if err := WriteSum(sumFile, sums); err != nil {
		return "", err
	}

	return version, nil
}

// resolveLatestStable runs git ls-remote --tags on the module URL and returns
// the highest stable semver tag. Returns NoStableVersionError if none found.
func resolveLatestStable(modPath string) (string, error) {
	url := gitURL(modPath)
	//nolint:gosec // modPath is from the parsed module manifest, not raw user input
	cmd := exec.Command(
		"git",
		"ls-remote",
		"--tags",
		url,
	)
	out, err := cmd.Output()
	if err != nil {
		return "", &AddError{
			Detail: "could not list tags for " + modPath + ": " + firstLine(out),
			Hint:   "check that " + url + " is accessible",
		}
	}

	version := ParseLatestStable(string(out))
	if version == "" {
		return "", &NoStableVersionError{ModPath: modPath}
	}
	return version, nil
}

// ParseLatestStable parses git ls-remote --tags output and returns the
// highest stable semver tag. Returns "" if no stable tags are found.
func ParseLatestStable(output string) string {
	var stable []string
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// Format: "<hash>\trefs/tags/<tag>"
		_, ref, ok := strings.Cut(line, "\t")
		if !ok {
			continue
		}
		// Skip dereferenced tags
		if strings.HasSuffix(ref, "^{}") {
			continue
		}
		tag, ok := strings.CutPrefix(ref, "refs/tags/")
		if !ok {
			continue
		}
		if !isSemver(tag) {
			continue
		}
		// Stable = no pre-release suffix
		rest := tag[1:] // strip 'v'
		if strings.ContainsRune(rest, '-') {
			continue
		}
		stable = append(stable, tag)
	}

	if len(stable) == 0 {
		return ""
	}

	slices.SortFunc(stable, compareSemver)
	return stable[len(stable)-1]
}

// compareSemver compares two semver strings for use with slices.SortFunc.
// Returns negative if a < b, zero if equal, positive if a > b.
func compareSemver(a, b string) int {
	pa := parseSemverParts(a)
	pb := parseSemverParts(b)
	for i := range pa {
		if pa[i] != pb[i] {
			return pa[i] - pb[i]
		}
	}
	return 0
}

// parseSemverParts extracts [major, minor, patch] from a semver string like "v1.2.3".
func parseSemverParts(v string) [3]int {
	rest := strings.TrimPrefix(v, "v")
	// Strip pre-release suffix
	if idx := strings.IndexByte(rest, '-'); idx >= 0 {
		rest = rest[:idx]
	}
	parts := strings.SplitN(rest, ".", 3)
	var out [3]int
	for i, p := range parts {
		if i >= 3 {
			break
		}
		n, _ := strconv.Atoi(p)
		out[i] = n
	}
	return out
}
