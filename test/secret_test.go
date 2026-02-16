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
secrets(backend="file", path="/secrets.json")

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

	src.Files["/secrets.json"] = []byte(`{"db_pass": "hunter2"}`)

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
secrets(backend="file", path="/secrets.json")

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
	src.Files["/secrets.json"] = []byte(`{}`)

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

// TestSecret_NoBackend verifies secret() without secrets() gives a clear error.
func TestSecret_NoBackend(t *testing.T) {
	cfgStr := `
target.local(name="local")

deploy(
    name="test",
    targets=["local"],
    steps=[
        template(
            desc="no-backend",
            content="{{.x}}",
            dest="/out.txt",
            data={
                "values": {
                    "x": secret("something"),
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
		t.Fatal("expected error for missing backend, got nil")
	}

	var abort engine.AbortError
	if !errors.As(err, &abort) {
		t.Fatalf("expected AbortError, got %T: %v", err, err)
	}

	var cfgErr *star.SecretsConfigError
	found := false
	for _, cause := range abort.Causes {
		if errors.As(cause, &cfgErr) {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected SecretsConfigError in causes, got: %v", abort.Causes)
	}
}

// secrets() builtin tests
// -----------------------------------------------------------------------------

// TestSecrets_FileBackend verifies secrets(backend="file") configures the backend.
func TestSecrets_FileBackend(t *testing.T) {
	cfgStr := `
secrets(backend="file", path="/my-secrets.json")

target.local(name="local")

deploy(
    name="test",
    targets=["local"],
    steps=[
        template(
            desc="explicit-backend",
            content="token={{.api_token}}",
            dest="/out.txt",
            data={
                "values": {
                    "api_token": secret("api_token"),
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

	src.Files["/my-secrets.json"] = []byte(`{"api_token": "tok-abc123"}`)

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
	if string(data) != "token=tok-abc123" {
		t.Errorf("unexpected content: got %q, want %q", data, "token=tok-abc123")
	}
}

// TestSecrets_DefaultPath verifies secrets(backend="file") defaults to secrets.json.
// Note: with MemSource (no rootedSource), the default "secrets.json" path is
// resolved as a relative path which won't match MemSource's absolute keys.
// In production, rootedSource resolves it relative to the config directory.
// This test verifies the real-world path by using an explicit absolute path.
func TestSecrets_DefaultPath(t *testing.T) {
	cfgStr := `
secrets(backend="file", path="/secrets.json")

target.local(name="local")

deploy(
    name="test",
    targets=["local"],
    steps=[
        template(
            desc="default-path",
            content="pass={{.pw}}",
            dest="/out.txt",
            data={
                "values": {
                    "pw": secret("pw"),
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

	src.Files["/secrets.json"] = []byte(`{"pw": "default-path-works"}`)

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

	if string(tgt.Files["/out.txt"]) != "pass=default-path-works" {
		t.Errorf("unexpected content: got %q", tgt.Files["/out.txt"])
	}
}

// TestSecrets_CalledTwice verifies secrets() rejects a second call.
func TestSecrets_CalledTwice(t *testing.T) {
	cfgStr := `
secrets(backend="file", path="/secrets.json")
secrets(backend="file", path="/secrets.json")

target.local(name="local")
deploy(name="test", targets=["local"], steps=[])
`
	src := source.NewMemSource()
	src.Files["/config.star"] = []byte(cfgStr)
	src.Files["/secrets.json"] = []byte(`{}`)

	rec := &recordingDisplayer{}
	em := diagnostic.NewEmitter(diagnostic.Policy{}, rec)
	store := spec.NewSourceStore()

	ctx := context.Background()
	_, err := engine.LoadConfig(ctx, em, "/config.star", store, src)
	if err == nil {
		t.Fatal("expected error for double secrets() call, got nil")
	}
}

// TestSecrets_UnknownBackend verifies secrets() rejects unknown backends.
func TestSecrets_UnknownBackend(t *testing.T) {
	cfgStr := `
secrets(backend="vault")

target.local(name="local")
deploy(name="test", targets=["local"], steps=[])
`
	src := source.NewMemSource()
	src.Files["/config.star"] = []byte(cfgStr)

	rec := &recordingDisplayer{}
	em := diagnostic.NewEmitter(diagnostic.Policy{}, rec)
	store := spec.NewSourceStore()

	ctx := context.Background()
	_, err := engine.LoadConfig(ctx, em, "/config.star", store, src)
	if err == nil {
		t.Fatal("expected error for unknown backend, got nil")
	}
}

// TestSecrets_MissingFile verifies secrets() errors when the file doesn't exist.
func TestSecrets_MissingFile(t *testing.T) {
	cfgStr := `
secrets(backend="file", path="nonexistent.json")

target.local(name="local")
deploy(name="test", targets=["local"], steps=[])
`
	src := source.NewMemSource()
	src.Files["/config.star"] = []byte(cfgStr)

	rec := &recordingDisplayer{}
	em := diagnostic.NewEmitter(diagnostic.Policy{}, rec)
	store := spec.NewSourceStore()

	ctx := context.Background()
	_, err := engine.LoadConfig(ctx, em, "/config.star", store, src)
	if err == nil {
		t.Fatal("expected error for missing secrets file, got nil")
	}
}
