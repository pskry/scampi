package engine

import (
	"context"
	"fmt"
	"time"

	"cuelang.org/go/cue/errors"
	"godoit.dev/doit/spec"
	"golang.org/x/sync/errgroup"
)

const taskTimeout = 5 * time.Second

func Apply(ctx context.Context, cfgPath string) error {
	reg, err := NewRegistry()
	if err != nil {
		return err
	}

	cfg, err := loadAndValidate(cfgPath, reg)
	if err != nil {
		errs := errors.Errors(err)
		fmt.Printf("CUE error summary:\n%v\n", err)
		fmt.Printf("CUE error details:\n%v\n", errors.Details(err, nil))
		fmt.Printf("CUE: %d error(s)\n", len(errs))
		return err
	}

	fmt.Printf("decoded config: %#v\n", cfg)
	p, err := plan(cfg)
	if err != nil {
		return err
	}

	return executePlan(ctx, p)
}

func plan(cfg spec.Config) (spec.RtPlan, error) {
	p := spec.RtPlan{}

	for i, t := range cfg.Tasks {
		rt, err := t.Spec.Plan(i, t.Config)
		if err != nil {
			return spec.RtPlan{}, err
		}
		p.Tasks = append(p.Tasks, rt)
	}

	return p, nil
}

func executePlan(ctx context.Context, plan spec.RtPlan) error {
	for _, task := range plan.Tasks {
		if err := executeTask(ctx, task); err != nil {
			return err
		}
	}

	return nil
}

func executeTask(ctx context.Context, task spec.RtTask) error {
	fmt.Printf("Running task: %s\n", task.Name())

	taskCtx, cancel := context.WithTimeout(ctx, taskTimeout)
	defer cancel()

	res, err := runTask(taskCtx, task)
	if err != nil {
		return fmt.Errorf("task %s failed: %w", task.Name(), err)
	}

	if res.Changed {
		fmt.Printf("Task %s changed state\n", task.Name())
	} else {
		fmt.Printf("Task %s already in desired state\n", task.Name())
	}

	return nil
}

func runTask(ctx context.Context, task spec.RtTask) (spec.Result, error) {
	grp, ctx := errgroup.WithContext(ctx)
	taskName := task.Name()
	ops := task.Ops()

	results := make([]spec.Result, len(ops))

	for i, op := range ops {
		fmt.Printf("Starting %s op: %s\n", taskName, op.Name())
		grp.Go(func() error {
			res, err := op.Execute(ctx)
			if err != nil {
				return err
			}

			results[i] = res
			fmt.Printf("Finished %s op: %s\n", taskName, op.Name())
			return nil
		})
	}

	if err := grp.Wait(); err != nil {
		return spec.Result{}, err
	}

	changed := false
	for _, res := range results {
		if res.Changed {
			changed = true
			break
		}
	}

	return spec.Result{Changed: changed}, nil
}
