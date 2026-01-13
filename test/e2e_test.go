package test

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"

	"godoit.dev/doit/diagnostic"
	"godoit.dev/doit/engine"
	"godoit.dev/doit/spec"
)

func TestCopyEndToEnd(t *testing.T) {
	tmp := t.TempDir()

	src := filepath.Join(tmp, "src.txt")
	dst := filepath.Join(tmp, "dst.txt")

	writeOrDie(src, []byte("hello"), 0o644)

	usr := currentUsr()
	cfg := fmt.Sprintf(`
package test

import "godoit.dev/doit/builtin"

units: [
	builtin.copy & {
		name:  "builtin.copy action"
		src:   %q
		dest:  %q
		perm:  "0644"
		owner: %q
		group: %q
	}
]
`, src, dst, usr.Uid, usr.Gid)

	cfgPath := filepath.Join(tmp, "config.cue")
	writeOrDie(cfgPath, []byte(cfg), 0o644)

	rec := &recordingDisplayer{}
	em := diagnostic.NewEmitter(diagnostic.Policy{}, rec)

	if err := engine.Apply(context.Background(), em, cfgPath, spec.NewSourceStore()); err != nil {
		t.Fatalf("expected successful call to engine.Apply, got err: %q\n%s", err, rec)
	}

	data := readOrDie(dst)
	if string(data) != "hello" {
		t.Fatalf("unexpected dest contents: %q", data)
	}
}
