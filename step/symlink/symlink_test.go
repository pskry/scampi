// SPDX-License-Identifier: GPL-3.0-only

package symlink

import (
	"path/filepath"
	"testing"
)

func TestResolveTarget(t *testing.T) {
	tests := []struct {
		name   string
		target string
		link   string
		want   string
	}{
		{
			name:   "absolute target unchanged",
			target: "/absolute/path/to/target",
			link:   "/some/link",
			want:   "/absolute/path/to/target",
		},
		{
			name:   "absolute target with absolute link",
			target: "/path/to/target.txt",
			link:   "/path/to/link.txt",
			want:   "/path/to/target.txt",
		},
		{
			name:   "relative target same directory",
			target: "./dir/target.txt",
			link:   "./dir/link.txt",
			want:   "target.txt",
		},
		{
			name:   "relative target parent directory",
			target: "./target.txt",
			link:   "./subdir/link.txt",
			want:   filepath.Join("..", "target.txt"),
		},
		{
			name:   "relative target sibling directory",
			target: "./other/target.txt",
			link:   "./subdir/link.txt",
			want:   filepath.Join("..", "other", "target.txt"),
		},
		{
			name:   "deeply nested relative paths",
			target: "./a/b/c/target.txt",
			link:   "./x/y/z/link.txt",
			want:   filepath.Join("..", "..", "..", "a", "b", "c", "target.txt"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := resolveTarget(tt.target, tt.link)
			if err != nil {
				t.Fatalf("resolveTarget() error = %v", err)
			}
			if got != tt.want {
				t.Errorf("resolveTarget() = %q, want %q", got, tt.want)
			}
		})
	}
}
