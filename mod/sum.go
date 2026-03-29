// SPDX-License-Identifier: GPL-3.0-only

package mod

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// ComputeHash computes a deterministic SHA-256 hash of the directory tree at dir.
// It skips .git/ directories, sorts paths lexicographically, and encodes each
// file as "<forward-slash-path>\0<content>" before hashing.
// Returns "h1:" + hex(sha256).
func ComputeHash(dir string) (string, error) {
	var paths []string

	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if d.Name() == ".git" {
				return filepath.SkipDir
			}
			return nil
		}
		rel, relErr := filepath.Rel(dir, path)
		if relErr != nil {
			return relErr
		}
		paths = append(paths, rel)
		return nil
	})
	if err != nil {
		return "", &SumError{
			Detail: "failed to walk directory: " + err.Error(),
			Hint:   "ensure the directory exists and is readable",
		}
	}

	sort.Strings(paths)

	h := sha256.New()
	for _, rel := range paths {
		slashPath := filepath.ToSlash(rel)
		data, readErr := os.ReadFile(filepath.Join(dir, rel))
		if readErr != nil {
			return "", &SumError{
				Detail: "failed to read file " + slashPath + ": " + readErr.Error(),
				Hint:   "ensure all files in the directory are readable",
			}
		}
		// sha256.Hash.Write never returns an error per the hash.Hash contract.
		//nolint:errcheck,revive
		h.Write([]byte(slashPath))
		//nolint:errcheck,revive
		h.Write([]byte{0})
		//nolint:errcheck,revive
		h.Write(data)
	}

	return "h1:" + hex.EncodeToString(h.Sum(nil)), nil
}

// ReadSum parses a scampi.sum file at path.
// Format per line: "<module> <version> <hash>".
// Key in the returned map is "<module> <version>".
// A missing file returns an empty map, not an error.
func ReadSum(path string) (map[string]string, error) {
	f, err := os.Open(path)
	if os.IsNotExist(err) {
		return map[string]string{}, nil
	}
	if err != nil {
		return nil, &SumError{
			Detail: "failed to open scampi.sum: " + err.Error(),
			Hint:   "ensure the file is readable",
		}
	}

	sums := map[string]string{}
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) != 3 {
			continue
		}
		key := fields[0] + " " + fields[1]
		sums[key] = fields[2]
	}
	closeErr := f.Close()
	if scanErr := scanner.Err(); scanErr != nil {
		return nil, &SumError{
			Detail: "failed to read scampi.sum: " + scanErr.Error(),
			Hint:   "ensure the file is not corrupted",
		}
	}
	if closeErr != nil {
		return nil, &SumError{
			Detail: "failed to close scampi.sum: " + closeErr.Error(),
			Hint:   "ensure the file is not corrupted",
		}
	}
	return sums, nil
}

// WriteSum writes sums to a scampi.sum file at path.
// Lines are sorted alphabetically by key and formatted as "<key> <hash>\n".
func WriteSum(path string, sums map[string]string) error {
	keys := make([]string, 0, len(sums))
	for k := range sums {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var sb strings.Builder
	for _, k := range keys {
		sb.WriteString(k)
		sb.WriteByte(' ')
		sb.WriteString(sums[k])
		sb.WriteByte('\n')
	}

	err := os.WriteFile(path, []byte(sb.String()), 0o644)
	if err != nil {
		return &SumError{
			Detail: "failed to write scampi.sum: " + err.Error(),
			Hint:   "ensure the directory is writable",
		}
	}
	return nil
}

// VerifyModule checks whether dep's hash in modDir matches what's recorded in sums.
// If dep is not yet in sums, nil is returned (new module, not yet recorded).
// If the hash mismatches, a SumMismatchError is returned.
func VerifyModule(dep Dependency, modDir string, sums map[string]string) error {
	key := dep.Path + " " + dep.Version
	expected, ok := sums[key]
	if !ok {
		return nil
	}

	actual, err := ComputeHash(modDir)
	if err != nil {
		return err
	}

	if actual != expected {
		return &SumMismatchError{
			ModPath:  dep.Path,
			Version:  dep.Version,
			Expected: expected,
			Actual:   actual,
		}
	}
	return nil
}
