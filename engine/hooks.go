// SPDX-License-Identifier: GPL-3.0-only

package engine

import (
	"context"
	"time"

	"scampi.dev/scampi/diagnostic"
	"scampi.dev/scampi/model"
)

// executeHooks runs notified hooks after all deploy steps complete.
// It collects which hooks were notified by steps that changed, then executes
// them in notification order. Hook chaining is supported: if a hook has
// on_change and it reports changes, those hooks are added to the queue.
// Each hook fires at most once per run.
func (e *Engine) executeHooks(
	ctx context.Context,
	stepReport model.ExecutionReport,
	hp *hookPlan,
	checkOnly bool,
	promisedPaths map[string]bool,
) (model.ExecutionReport, error) {
	if hp == nil || len(hp.actions) == 0 {
		return model.ExecutionReport{}, nil
	}

	// Collect notified hooks from step results, preserving notification order.
	var queue []string
	notified := map[string]bool{}
	triggerBy := map[string]string{} // hook ID → desc of step that triggered it

	for i, ar := range stepReport.Actions {
		onChange, ok := hp.onChange[i]
		if !ok {
			continue
		}

		changed := actionChanged(ar, checkOnly)
		if !changed {
			continue
		}

		for _, hookID := range onChange {
			if !notified[hookID] {
				notified[hookID] = true
				triggerBy[hookID] = ar.Action.Desc()
				queue = append(queue, hookID)
			}
		}
	}

	// Execute notified hooks. Process queue — new entries may be appended
	// by hook chaining.
	var hookReports []model.ActionReport
	executed := map[string]bool{}

	for i := 0; i < len(queue); i++ {
		hookID := queue[i]
		if executed[hookID] {
			continue
		}
		executed[hookID] = true

		actions, ok := hp.actions[hookID]
		if !ok {
			continue
		}

		hookStart := time.Now()
		var aggregate model.ActionSummary
		anyChanged := false
		var hookErr error

		for _, act := range actions {
			hookIdx := len(stepReport.Actions) + len(hookReports)

			actCtx, cancel := context.WithTimeout(ctx, actionTimeout)

			var ar model.ActionReport
			var err error
			if checkOnly {
				ar, err = e.runCheckAction(actCtx, hookIdx, act, promisedPaths, hookID)
			} else {
				ar, err = e.runAction(actCtx, hookIdx, act, hookID)
			}
			cancel()

			hookReports = append(hookReports, ar)
			addSummary(&aggregate, ar.Summary)

			if actionChanged(ar, checkOnly) {
				anyChanged = true
			}

			if err != nil {
				hookErr = err
				break
			}
		}

		e.em.EmitActionLifecycle(diagnostic.HookTriggered(hookID, triggerBy[hookID], aggregate, time.Since(hookStart)))

		if hookErr != nil {
			return model.ExecutionReport{
				Actions: hookReports,
				Err:     hookErr,
			}, hookErr
		}

		// Handle chaining: if any action in this hook changed, notify on_change targets
		if anyChanged {
			if steps, ok := e.cfg.Hooks[hookID]; ok {
				for _, step := range steps {
					for _, nextID := range step.OnChange {
						if !notified[nextID] {
							notified[nextID] = true
							triggerBy[nextID] = "hook:" + hookID
							queue = append(queue, nextID)
						}
					}
				}
			}
		}
	}

	// Emit HookSkipped for hooks that were defined but never notified
	for id := range hp.actions {
		if !notified[id] {
			e.em.EmitActionLifecycle(diagnostic.HookSkipped(id))
		}
	}

	return model.ExecutionReport{Actions: hookReports}, nil
}

// addSummary accumulates src into dst.
func addSummary(dst *model.ActionSummary, src model.ActionSummary) {
	dst.Total += src.Total
	dst.Succeeded += src.Succeeded
	dst.Failed += src.Failed
	dst.Aborted += src.Aborted
	dst.Skipped += src.Skipped
	dst.Changed += src.Changed
	dst.WouldChange += src.WouldChange
}

// actionChanged returns true if an action report indicates something changed
// (or would change in check mode).
func actionChanged(ar model.ActionReport, checkOnly bool) bool {
	if checkOnly {
		return ar.Summary.WouldChange > 0
	}
	return ar.Summary.Changed > 0
}
