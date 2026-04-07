// SPDX-License-Identifier: GPL-3.0-only

package std

import (
	"io/fs"
	"strings"
	"testing"

	"scampi.dev/scampi/lang/check"
	"scampi.dev/scampi/lang/lex"
	"scampi.dev/scampi/lang/parse"
)

func TestStdLibCompiles(t *testing.T) {
	t.Skip("submodule stubs reference cross-module types; needs registry-driven type resolution")
	err := fs.WalkDir(FS, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(d.Name(), ".scampi") {
			return nil
		}
		t.Run(path, func(t *testing.T) {
			data, err := fs.ReadFile(FS, path)
			if err != nil {
				t.Fatalf("read: %v", err)
			}

			l := lex.New(path, data)
			p := parse.New(l)
			f := p.Parse()

			if errs := l.Errors(); len(errs) > 0 {
				t.Fatalf("lex errors: %v", errs)
			}
			if errs := p.Errors(); len(errs) > 0 {
				t.Fatalf("parse errors: %v", errs)
			}

			c := check.New()
			c.Check(f)
			if errs := c.Errors(); len(errs) > 0 {
				for _, e := range errs {
					t.Errorf("check: %s", e.Msg)
				}
			}
		})
		return nil
	})
	if err != nil {
		t.Fatalf("walk: %v", err)
	}
}
