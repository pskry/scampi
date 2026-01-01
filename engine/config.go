package engine

import (
	"fmt"
	"reflect"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	"cuelang.org/go/cue/load"
	"godoit.dev/doit/spec"
)

func loadAndValidate(cfgPath string, reg *Registry) (spec.Config, error) {
	ctx := cuecontext.New()

	cfg, err := loadConfig(ctx, cfgPath)
	if err != nil {
		return spec.Config{}, err
	}

	return decodeConfig(cfg, reg)
}

func loadConfig(ctx *cue.Context, path string) (cue.Value, error) {
	cfg := &load.Config{
		// CRITICAL: module root
		Dir: ".", // repo root (where cue.mod lives)
	}

	instances := load.Instances([]string{path}, cfg)
	if len(instances) == 0 {
		return cue.Value{}, fmt.Errorf("no CUE instances loaded")
	}

	if err := instances[0].Err; err != nil {
		return cue.Value{}, err
	}

	val := ctx.BuildInstance(instances[0])
	if err := val.Err(); err != nil {
		return cue.Value{}, err
	}

	return val, nil
}

func decodeConfig(configVal cue.Value, reg *Registry) (spec.Config, error) {
	tasksVal := configVal.LookupPath(cue.ParsePath("playbook.tasks"))
	if err := tasksVal.Err(); err != nil {
		return spec.Config{}, err
	}

	iter, err := tasksVal.List()
	if err != nil {
		return spec.Config{}, err
	}

	cfg := spec.Config{}
	for iter.Next() {
		taskVal := iter.Value()

		var kind string
		if err := taskVal.LookupPath(cue.ParsePath("kind")).Decode(&kind); err != nil {
			return spec.Config{}, err
		}

		s, ok := reg.SpecForKind(kind)
		if !ok {
			return spec.Config{}, fmt.Errorf("task kind %s not found in schema", kind)
		}

		out := s.NewConfig()

		if !isPointer(out) {
			return spec.Config{}, fmt.Errorf("spec['%s'].NewConfig must return a pointer. Got %T", kind, out)
		}

		if err := decodeTask(taskVal, kind, out); err != nil {
			return spec.Config{}, err
		}

		cfg.Tasks = append(cfg.Tasks, spec.CfgTask{
			Kind:   kind,
			Spec:   s,
			Config: out,
		})
	}

	return cfg, nil
}

func decodeTask(taskVal cue.Value, kind string, out any) error {
	specVal := taskVal.LookupPath(cue.ParsePath(kind))
	if !specVal.Exists() {
		return fmt.Errorf("task of kind %q is missing required %q block", kind, kind)
	}

	if err := specVal.Decode(out); err != nil {
		return fmt.Errorf("failed to decode %q task config: %w", kind, err)
	}

	return nil
}

func isPointer(i any) bool {
	v := reflect.ValueOf(i)
	return v.Kind() == reflect.Pointer
}
