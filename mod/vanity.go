// SPDX-License-Identifier: GPL-3.0-only

package mod

import (
	"context"
	"io"
	"net/http"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// resolvedModule is the result of resolving a module path to a git
// clone URL and an optional subdirectory within the repository.
type resolvedModule struct {
	URL    string
	Subdir string // empty when module is at repo root
}

// resolveModule returns the clone URL (and optional subdir) for a
// module path. It walks the path from longest to shortest, probing
// each candidate as a git repo. If no repo is found, it falls back
// to vanity import resolution via the scampi-get meta tag.
func resolveModule(modPath string) resolvedModule {
	if filepath.IsAbs(modPath) {
		return resolvedModule{URL: modPath}
	}

	// Try progressively shorter paths as repo roots.
	// A valid module path has at least 3 segments (host/org/repo).
	segments := strings.Split(modPath, "/")
	for i := len(segments); i >= 3; i-- {
		candidate := strings.Join(segments[:i], "/")
		url := "https://" + candidate + ".git"
		if probeGitRepo(url) {
			subdir := ""
			if i < len(segments) {
				subdir = strings.Join(segments[i:], "/")
			}
			return resolvedModule{URL: url, Subdir: subdir}
		}
	}

	// No git repo found — try vanity import resolution.
	if rm, ok := resolveVanityImport(modPath); ok {
		return rm
	}

	// Last resort: assume full path is the repo. This will fail at
	// clone time with a useful error from git.
	return resolvedModule{URL: "https://" + modPath + ".git"}
}

// probeGitRepo checks whether a git remote URL points to a valid
// repository by running git ls-remote.
func probeGitRepo(url string) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "git", "ls-remote", "--heads", url)
	cmd.Stdout = io.Discard
	cmd.Stderr = io.Discard
	return cmd.Run() == nil
}

// resolveVanityImport fetches https://<modPath>?scampi-get=1 and looks
// for a <meta name="scampi-import" content="<prefix> git <url>"> tag.
func resolveVanityImport(modPath string) (resolvedModule, bool) {
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get("https://" + modPath + "?scampi-get=1")
	if err != nil {
		return resolvedModule{}, false
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return resolvedModule{}, false
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return resolvedModule{}, false
	}

	return parseScampiImportMeta(string(body), modPath)
}

// parseScampiImportMeta extracts the clone URL and subdirectory from a
// scampi-import meta tag. The expected format is:
//
//	<meta name="scampi-import" content="<prefix> git <repo-url>">
//
// The prefix must match or be a prefix of modPath. When the module path
// extends beyond the prefix (e.g. prefix "scampi.dev/modules", modPath
// "scampi.dev/modules/npm"), the remainder becomes the Subdir ("npm").
func parseScampiImportMeta(
	html, modPath string,
) (resolvedModule, bool) {
	const marker = `name="scampi-import"`

	for {
		idx := strings.Index(html, marker)
		if idx < 0 {
			return resolvedModule{}, false
		}
		tagStart := strings.LastIndex(html[:idx], "<")
		if tagStart < 0 {
			html = html[idx+len(marker):]
			continue
		}
		tagEnd := strings.Index(html[tagStart:], ">")
		if tagEnd < 0 {
			html = html[idx+len(marker):]
			continue
		}
		tag := html[tagStart : tagStart+tagEnd+1]

		content := extractAttr(tag, "content")
		if content == "" {
			html = html[idx+len(marker):]
			continue
		}

		fields := strings.Fields(content)
		if len(fields) != 3 {
			html = html[idx+len(marker):]
			continue
		}

		prefix, vcs, repoURL := fields[0], fields[1], fields[2]
		if vcs != "git" {
			html = html[idx+len(marker):]
			continue
		}

		if modPath == prefix {
			return resolvedModule{URL: repoURL}, true
		}
		if strings.HasPrefix(modPath, prefix+"/") {
			subdir := modPath[len(prefix)+1:]
			return resolvedModule{URL: repoURL, Subdir: subdir}, true
		}

		html = html[idx+len(marker):]
	}
}

// extractAttr extracts the value of an HTML attribute from a tag string.
func extractAttr(tag, attr string) string {
	key := attr + `="`
	idx := strings.Index(tag, key)
	if idx < 0 {
		return ""
	}
	start := idx + len(key)
	end := strings.Index(tag[start:], `"`)
	if end < 0 {
		return ""
	}
	return tag[start : start+end]
}
