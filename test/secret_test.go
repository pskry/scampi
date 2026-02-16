// SPDX-License-Identifier: GPL-3.0-only

package test

import (
	"context"
	"errors"
	"testing"

	"godoit.dev/doit/diagnostic"
	"godoit.dev/doit/engine"
	"godoit.dev/doit/source"
	"godoit.dev/doit/spec"
	"godoit.dev/doit/star"
	"godoit.dev/doit/target"
)

// TestSecret_ResolvesIntoTemplateData verifies that secret() values flow
// through to template rendering.
func TestSecret_ResolvesIntoTemplateData(t *testing.T) {
	cfgStr := `
target.local(name="local")

deploy(
    name="test",
    targets=["local"],
    steps=[
        template(
            desc="secret-template",
            content="pass={{.db_pass}}",
            dest="/out.txt",
            data={
                "values": {
                    "db_pass": secret("db_pass"),
                },
            },
            perm="0644",
            owner="user",
            group="group",
        ),
    ],
)
`
	src := source.NewMemSource()
	tgt := target.NewMemTarget()

	src.Secrets["db_pass"] = "hunter2"

	rec := &recordingDisplayer{}
	em := diagnostic.NewEmitter(diagnostic.Policy{}, rec)
	store := spec.NewSourceStore()

	e, err := loadAndResolve(t, cfgStr, src, tgt, em, store)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer e.Close()

	err = e.Apply(context.Background())
	if err != nil {
		t.Fatalf("Apply failed: %v\n%s", err, rec)
	}

	data, ok := tgt.Files["/out.txt"]
	if !ok {
		t.Fatal("destination file not created")
	}
	if string(data) != "pass=hunter2" {
		t.Errorf("unexpected content: got %q, want %q", data, "pass=hunter2")
	}
}

// TestSecret_NotFound verifies that a missing secret produces an abort.
func TestSecret_NotFound(t *testing.T) {
	cfgStr := `
target.local(name="local")

deploy(
    name="test",
    targets=["local"],
    steps=[
        template(
            desc="missing-secret",
            content="{{.token}}",
            dest="/out.txt",
            data={
                "values": {
                    "token": secret("missing_key"),
                },
            },
            perm="0644",
            owner="user",
            group="group",
        ),
    ],
)
`
	src := source.NewMemSource()
	src.Files["/config.star"] = []byte(cfgStr)

	rec := &recordingDisplayer{}
	em := diagnostic.NewEmitter(diagnostic.Policy{}, rec)
	store := spec.NewSourceStore()

	ctx := context.Background()
	_, err := engine.LoadConfig(ctx, em, "/config.star", store, src)
	if err == nil {
		t.Fatal("expected error for missing secret, got nil")
	}

	var abort engine.AbortError
	if !errors.As(err, &abort) {
		t.Fatalf("expected AbortError, got %T: %v", err, err)
	}

	var notFound *star.SecretNotFoundError
	found := false
	for _, cause := range abort.Causes {
		if errors.As(cause, &notFound) {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected SecretNotFoundError in causes, got: %v", abort.Causes)
	}
}

// TestSecret_WrongArgType verifies secret() rejects non-string keys.
func TestSecret_WrongArgType(t *testing.T) {
	cfgStr := `
target.local(name="local")

deploy(
    name="test",
    targets=["local"],
    steps=[
        template(
            desc="bad-secret",
            content="{{.x}}",
            dest="/out.txt",
            data={
                "values": {
                    "x": secret(42),
                },
            },
            perm="0644",
            owner="user",
            group="group",
        ),
    ],
)
`
	src := source.NewMemSource()
	src.Files["/config.star"] = []byte(cfgStr)

	rec := &recordingDisplayer{}
	em := diagnostic.NewEmitter(diagnostic.Policy{}, rec)
	store := spec.NewSourceStore()

	ctx := context.Background()
	_, err := engine.LoadConfig(ctx, em, "/config.star", store, src)
	if err == nil {
		t.Fatal("expected error for wrong arg type, got nil")
	}

	var abort engine.AbortError
	if !errors.As(err, &abort) {
		t.Fatalf("expected AbortError, got %T: %v", err, err)
	}

	var secretErr *star.SecretError
	found := false
	for _, cause := range abort.Causes {
		if errors.As(cause, &secretErr) {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected SecretError in causes, got: %v", abort.Causes)
	}
}

// TestSecret_TooManyArgs verifies secret() rejects extra arguments.
func TestSecret_TooManyArgs(t *testing.T) {
	cfgStr := `
target.local(name="local")

deploy(
    name="test",
    targets=["local"],
    steps=[
        template(
            desc="too-many",
            content="{{.x}}",
            dest="/out.txt",
            data={
                "values": {
                    "x": secret("a", "b"),
                },
            },
            perm="0644",
            owner="user",
            group="group",
        ),
    ],
)
`
	src := source.NewMemSource()
	src.Files["/config.star"] = []byte(cfgStr)

	rec := &recordingDisplayer{}
	em := diagnostic.NewEmitter(diagnostic.Policy{}, rec)
	store := spec.NewSourceStore()

	ctx := context.Background()
	_, err := engine.LoadConfig(ctx, em, "/config.star", store, src)
	if err == nil {
		t.Fatal("expected error for too many args, got nil")
	}
}
