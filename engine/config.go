package engine

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"reflect"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	"cuelang.org/go/cue/load"
	"godoit.dev/doit"
	"godoit.dev/doit/spec"
)

type overlayFS struct {
	Embedded fs.FS
	Host     fs.FS
}

func (o overlayFS) Open(name string) (fs.File, error) {
	f, err := o.Embedded.Open(name)
	if err == nil {
		return f, nil
	}
	if errors.Is(err, fs.ErrNotExist) {
		s, err := o.Host.Open(name)
		if err == nil {
			return s, err
		}

		return s, err
	}

	return nil, err
}

func loadAndValidate(cfgPath string, reg *Registry) (spec.Config, error) {
	ctx := cuecontext.New()

	cfg, err := loadConfig(ctx, cfgPath)
	if err != nil {
		return spec.Config{}, err
	}

	return decodeConfig(cfg, reg)
}

func loadConfig(ctx *cue.Context, path string) (cue.Value, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return cue.Value{}, err
	}

	emb, err := fs.Sub(doit.EmbeddedSchemaModule, "cue")
	if err != nil {
		return cue.Value{}, err
	}

	// One loader config for both schema and user config
	loaderCfg := &load.Config{
		FS: overlayFS{
			Embedded: emb,
			Host:     os.DirFS(cwd),
		},
		Dir: ".",
	}

	// --- load user config ---
	userInstances := load.Instances([]string{path}, loaderCfg)
	if len(userInstances) == 0 {
		return cue.Value{}, fmt.Errorf("no user config instances loaded")
	}
	if err := userInstances[0].Err; err != nil {
		return cue.Value{}, err
	}

	userVal := ctx.BuildInstance(userInstances[0])
	if err := userVal.Err(); err != nil {
		return cue.Value{}, err
	}

	schemaInstances := load.Instances([]string{"godoit.dev/doit/core"}, loaderCfg)
	if len(schemaInstances) == 0 {
		return cue.Value{}, fmt.Errorf("no schema instances loaded")
	}
	if err := schemaInstances[0].Err; err != nil {
		return cue.Value{}, err
	}

	schemaPkg := ctx.BuildInstance(schemaInstances[0])
	if err := schemaPkg.Err(); err != nil {
		return cue.Value{}, err
	}

	// --- apply schema ---
	val := schemaPkg.Value().Unify(userVal)
	if err := val.Err(); err != nil {
		return cue.Value{}, err
	}

	return val, nil
}

func decodeConfig(configVal cue.Value, reg *Registry) (spec.Config, error) {
	tasksVal := configVal.LookupPath(cue.ParsePath("tasks"))
	if err := tasksVal.Err(); err != nil {
		return spec.Config{}, err
	}

	iter, err := tasksVal.Fields()
	if err != nil {
		return spec.Config{}, err
	}

	cfg := spec.Config{}
	for iter.Next() {
		name := iter.Selector().String()
		taskVal := iter.Value()

		metaVal := taskVal.LookupPath(cue.ParsePath("meta"))
		if err := metaVal.Err(); err != nil {
			return spec.Config{}, err
		}

		kindVal := metaVal.LookupPath(cue.ParsePath("kind"))
		if err := kindVal.Err(); err != nil {
			return spec.Config{}, err
		}

		kind, err := kindVal.String()
		if err != nil {
			return spec.Config{}, err
		}

		s, ok := reg.SpecForKind(kind)
		if !ok {
			return spec.Config{}, fmt.Errorf("unknown task kind %q", kind)
		}
		c := s.NewConfig()
		if !isPointer(c) {
			return spec.Config{}, fmt.Errorf("spec['%s'].NewConfig must return a pointer. Got %T", s.Kind(), c)
		}

		if err := taskVal.Decode(c); err != nil {
			return spec.Config{}, err
		}

		cfg.Tasks = append(cfg.Tasks, spec.CfgTask{
			Name:   name,
			Spec:   s,
			Config: c,
		})
	}

	return cfg, nil
}

func isPointer(i any) bool {
	v := reflect.ValueOf(i)
	return v.Kind() == reflect.Pointer
}
