// SPDX-License-Identifier: GPL-3.0-only

package star_test

import (
	"context"
	"testing"

	"godoit.dev/doit/source"
	"godoit.dev/doit/spec"
	"godoit.dev/doit/star"
	stepcopy "godoit.dev/doit/step/copy"
	"godoit.dev/doit/step/dir"
	"godoit.dev/doit/step/pkg"
	"godoit.dev/doit/step/symlink"
	"godoit.dev/doit/step/template"
	"godoit.dev/doit/target/local"
	"godoit.dev/doit/target/ssh"
)

func TestEvalMinimalConfig(t *testing.T) {
	src := source.NewMemSource()
	src.Files["/config.star"] = []byte(`
target.local(name="myhost")

deploy(
    name="main",
    targets=["myhost"],
    steps=[
        copy(src="./file.txt", dest="/tmp/file.txt", perm="0644", owner="root", group="root"),
    ],
)
`)

	store := spec.NewSourceStore()
	cfg, err := star.Eval(context.Background(), "/config.star", store, src)
	if err != nil {
		t.Fatalf("Eval failed: %v", err)
	}

	if cfg.Path != "/config.star" {
		t.Errorf("Path = %q, want /config.star", cfg.Path)
	}

	if len(cfg.Targets) != 1 {
		t.Fatalf("got %d targets, want 1", len(cfg.Targets))
	}
	tgt, ok := cfg.Targets["myhost"]
	if !ok {
		t.Fatal("target 'myhost' not found")
	}
	if tgt.Type.Kind() != "local" {
		t.Errorf("target kind = %q, want 'local'", tgt.Type.Kind())
	}
	if _, ok := tgt.Config.(*local.Config); !ok {
		t.Errorf("target config type = %T, want *local.Config", tgt.Config)
	}

	if len(cfg.Deploy) != 1 {
		t.Fatalf("got %d deploy blocks, want 1", len(cfg.Deploy))
	}
	block, ok := cfg.Deploy["main"]
	if !ok {
		t.Fatal("deploy 'main' not found")
	}
	if len(block.Targets) != 1 || block.Targets[0] != "myhost" {
		t.Errorf("deploy targets = %v, want [myhost]", block.Targets)
	}
	if len(block.Steps) != 1 {
		t.Fatalf("got %d steps, want 1", len(block.Steps))
	}

	step := block.Steps[0]
	if step.Type.Kind() != "copy" {
		t.Errorf("step kind = %q, want 'copy'", step.Type.Kind())
	}
	cc, ok := step.Config.(*stepcopy.CopyConfig)
	if !ok {
		t.Fatalf("step config type = %T, want *copy.CopyConfig", step.Config)
	}
	if cc.Src != "./file.txt" {
		t.Errorf("copy.Src = %q, want './file.txt'", cc.Src)
	}
	if cc.Dest != "/tmp/file.txt" {
		t.Errorf("copy.Dest = %q, want '/tmp/file.txt'", cc.Dest)
	}
}

func TestEvalAllStepTypes(t *testing.T) {
	src := source.NewMemSource()
	src.Files["/config.star"] = []byte(`
target.local(name="host")

steps = [
    copy(src="./a", dest="/tmp/a", perm="0644", owner="u", group="g"),
    dir(path="/tmp/d", perm="0755", owner="u", group="g"),
    pkg(packages=["curl", "wget"]),
    symlink(target="/tmp/a", link="/tmp/l"),
    template(dest="/tmp/t", perm="0644", owner="u", group="g", content="hello"),
]

deploy(name="main", targets=["host"], steps=steps)
`)

	store := spec.NewSourceStore()
	cfg, err := star.Eval(context.Background(), "/config.star", store, src)
	if err != nil {
		t.Fatalf("Eval failed: %v", err)
	}

	block := cfg.Deploy["main"]
	if len(block.Steps) != 5 {
		t.Fatalf("got %d steps, want 5", len(block.Steps))
	}

	kinds := []string{"copy", "dir", "pkg", "symlink", "template"}
	for i, want := range kinds {
		if got := block.Steps[i].Type.Kind(); got != want {
			t.Errorf("step[%d] kind = %q, want %q", i, got, want)
		}
	}

	// Verify pkg config
	pc := block.Steps[2].Config.(*pkg.PkgConfig)
	if len(pc.Packages) != 2 || pc.Packages[0] != "curl" || pc.Packages[1] != "wget" {
		t.Errorf("pkg.Packages = %v, want [curl wget]", pc.Packages)
	}
	if pc.State != "present" {
		t.Errorf("pkg.State = %q, want 'present'", pc.State)
	}
}

func TestEvalSSHTarget(t *testing.T) {
	src := source.NewMemSource()
	src.Files["/config.star"] = []byte(`
target.ssh(
    name="remote",
    host="192.168.1.1",
    user="admin",
    port=2222,
    key="~/.ssh/id_rsa",
    insecure=True,
    timeout="10s",
)

deploy(name="main", targets=["remote"], steps=[
    dir(path="/tmp/test"),
])
`)

	store := spec.NewSourceStore()
	cfg, err := star.Eval(context.Background(), "/config.star", store, src)
	if err != nil {
		t.Fatalf("Eval failed: %v", err)
	}

	tgt, ok := cfg.Targets["remote"]
	if !ok {
		t.Fatal("target 'remote' not found")
	}
	if tgt.Type.Kind() != "ssh" {
		t.Errorf("target kind = %q, want 'ssh'", tgt.Type.Kind())
	}

	sc, ok := tgt.Config.(*ssh.Config)
	if !ok {
		t.Fatalf("target config type = %T, want *ssh.Config", tgt.Config)
	}
	if sc.Host != "192.168.1.1" {
		t.Errorf("ssh.Host = %q, want '192.168.1.1'", sc.Host)
	}
	if sc.User != "admin" {
		t.Errorf("ssh.User = %q, want 'admin'", sc.User)
	}
	if sc.Port != 2222 {
		t.Errorf("ssh.Port = %d, want 2222", sc.Port)
	}
	if sc.Key != "~/.ssh/id_rsa" {
		t.Errorf("ssh.Key = %q, want '~/.ssh/id_rsa'", sc.Key)
	}
	if !sc.Insecure {
		t.Error("ssh.Insecure = false, want true")
	}
	if sc.Timeout != "10s" {
		t.Errorf("ssh.Timeout = %q, want '10s'", sc.Timeout)
	}
}

func TestEvalEnvDefault(t *testing.T) {
	src := source.NewMemSource()
	src.Env["MY_HOST"] = "remotehost"
	src.Env["MY_PORT"] = "3000"
	src.Files["/config.star"] = []byte(`
host = env("MY_HOST", "localhost")
port = env("MY_PORT", 22)
missing = env("NOPE", "fallback")

target.ssh(name="t", host=host, user="admin", port=port)
deploy(name="main", targets=["t"], steps=[
    dir(path="/tmp/" + missing),
])
`)

	store := spec.NewSourceStore()
	cfg, err := star.Eval(context.Background(), "/config.star", store, src)
	if err != nil {
		t.Fatalf("Eval failed: %v", err)
	}

	sc := cfg.Targets["t"].Config.(*ssh.Config)
	if sc.Host != "remotehost" {
		t.Errorf("host = %q, want 'remotehost'", sc.Host)
	}
	if sc.Port != 3000 {
		t.Errorf("port = %d, want 3000", sc.Port)
	}

	step := cfg.Deploy["main"].Steps[0]
	dc := step.Config.(*dir.DirConfig)
	if dc.Path != "/tmp/fallback" {
		t.Errorf("path = %q, want '/tmp/fallback'", dc.Path)
	}
}

func TestEvalEnvRequired(t *testing.T) {
	src := source.NewMemSource()
	src.Files["/config.star"] = []byte(`
val = env("REQUIRED_VAR")
`)

	store := spec.NewSourceStore()
	_, err := star.Eval(context.Background(), "/config.star", store, src)
	if err == nil {
		t.Fatal("expected error for missing required env var")
	}
}

func TestEvalDuplicateTarget(t *testing.T) {
	src := source.NewMemSource()
	src.Files["/config.star"] = []byte(`
target.local(name="dup")
target.local(name="dup")
`)

	store := spec.NewSourceStore()
	_, err := star.Eval(context.Background(), "/config.star", store, src)
	if err == nil {
		t.Fatal("expected error for duplicate target")
	}
}

func TestEvalDuplicateDeploy(t *testing.T) {
	src := source.NewMemSource()
	src.Files["/config.star"] = []byte(`
target.local(name="host")
deploy(name="main", targets=["host"], steps=[])
deploy(name="main", targets=["host"], steps=[])
`)

	store := spec.NewSourceStore()
	_, err := star.Eval(context.Background(), "/config.star", store, src)
	if err == nil {
		t.Fatal("expected error for duplicate deploy block")
	}
}

func TestEvalStepComposition(t *testing.T) {
	src := source.NewMemSource()
	src.Files["/config.star"] = []byte(`
target.local(name="host")

base = [
    dir(path="/tmp/a"),
    dir(path="/tmp/b"),
]
extra = [
    dir(path="/tmp/c"),
]

deploy(name="main", targets=["host"], steps=base + extra)
`)

	store := spec.NewSourceStore()
	cfg, err := star.Eval(context.Background(), "/config.star", store, src)
	if err != nil {
		t.Fatalf("Eval failed: %v", err)
	}

	block := cfg.Deploy["main"]
	if len(block.Steps) != 3 {
		t.Fatalf("got %d steps, want 3", len(block.Steps))
	}

	paths := []string{"/tmp/a", "/tmp/b", "/tmp/c"}
	for i, want := range paths {
		dc := block.Steps[i].Config.(*dir.DirConfig)
		if dc.Path != want {
			t.Errorf("step[%d] path = %q, want %q", i, dc.Path, want)
		}
	}
}

func TestEvalTemplateWithData(t *testing.T) {
	src := source.NewMemSource()
	src.Files["/config.star"] = []byte(`
target.local(name="host")

deploy(name="main", targets=["host"], steps=[
    template(
        dest="/tmp/out",
        perm="0644",
        owner="root",
        group="root",
        content="hello {{ .Name }}",
        data={
            "values": {"Name": "world"},
            "env": {"HOME": "/root"},
        },
    ),
])
`)

	store := spec.NewSourceStore()
	cfg, err := star.Eval(context.Background(), "/config.star", store, src)
	if err != nil {
		t.Fatalf("Eval failed: %v", err)
	}

	step := cfg.Deploy["main"].Steps[0]
	tc := step.Config.(*template.TemplateConfig)
	if tc.Content != "hello {{ .Name }}" {
		t.Errorf("content = %q", tc.Content)
	}
	if tc.Data.Values["Name"] != "world" {
		t.Errorf("data.values.Name = %v, want 'world'", tc.Data.Values["Name"])
	}
	if tc.Data.Env["HOME"] != "/root" {
		t.Errorf("data.env.HOME = %v, want '/root'", tc.Data.Env["HOME"])
	}
}

func TestEvalLoad(t *testing.T) {
	src := source.NewMemSource()
	src.Files["/lib/steps.star"] = []byte(`
common = [
    dir(path="/tmp/shared"),
]
`)
	src.Files["/config.star"] = []byte(`
load("/lib/steps.star", "common")

target.local(name="host")
deploy(name="main", targets=["host"], steps=common)
`)

	store := spec.NewSourceStore()
	cfg, err := star.Eval(context.Background(), "/config.star", store, src)
	if err != nil {
		t.Fatalf("Eval failed: %v", err)
	}

	block := cfg.Deploy["main"]
	if len(block.Steps) != 1 {
		t.Fatalf("got %d steps, want 1", len(block.Steps))
	}
	dc := block.Steps[0].Config.(*dir.DirConfig)
	if dc.Path != "/tmp/shared" {
		t.Errorf("path = %q, want '/tmp/shared'", dc.Path)
	}
}

func TestEvalSourceSpans(t *testing.T) {
	src := source.NewMemSource()
	src.Files["/config.star"] = []byte(`target.local(name="host")
deploy(name="main", targets=["host"], steps=[
    dir(path="/tmp/x"),
])
`)

	store := spec.NewSourceStore()
	cfg, err := star.Eval(context.Background(), "/config.star", store, src)
	if err != nil {
		t.Fatalf("Eval failed: %v", err)
	}

	step := cfg.Deploy["main"].Steps[0]
	if step.Source.Filename != "/config.star" {
		t.Errorf("source filename = %q, want '/config.star'", step.Source.Filename)
	}
	if step.Source.StartLine == 0 {
		t.Error("source line = 0, expected non-zero")
	}
}

func TestEvalPkgStateAbsent(t *testing.T) {
	src := source.NewMemSource()
	src.Files["/config.star"] = []byte(`
target.local(name="host")
deploy(name="main", targets=["host"], steps=[
    pkg(packages=["vim"], state="absent"),
])
`)

	store := spec.NewSourceStore()
	cfg, err := star.Eval(context.Background(), "/config.star", store, src)
	if err != nil {
		t.Fatalf("Eval failed: %v", err)
	}

	pc := cfg.Deploy["main"].Steps[0].Config.(*pkg.PkgConfig)
	if pc.State != "absent" {
		t.Errorf("state = %q, want 'absent'", pc.State)
	}
}

func TestEvalSymlinkConfig(t *testing.T) {
	src := source.NewMemSource()
	src.Files["/config.star"] = []byte(`
target.local(name="host")
deploy(name="main", targets=["host"], steps=[
    symlink(target="/usr/bin/vim", link="/usr/local/bin/vim", desc="vim link"),
])
`)

	store := spec.NewSourceStore()
	cfg, err := star.Eval(context.Background(), "/config.star", store, src)
	if err != nil {
		t.Fatalf("Eval failed: %v", err)
	}

	step := cfg.Deploy["main"].Steps[0]
	sc := step.Config.(*symlink.SymlinkConfig)
	if sc.Target != "/usr/bin/vim" {
		t.Errorf("target = %q", sc.Target)
	}
	if sc.Link != "/usr/local/bin/vim" {
		t.Errorf("link = %q", sc.Link)
	}
	if sc.Desc != "vim link" {
		t.Errorf("desc = %q", sc.Desc)
	}
}

// Suppress unused import warnings
var (
	_ = stepcopy.Copy{}
	_ = dir.Dir{}
	_ = pkg.Pkg{}
	_ = symlink.Symlink{}
	_ = template.Template{}
	_ = local.Local{}
	_ = ssh.SSH{}
)
