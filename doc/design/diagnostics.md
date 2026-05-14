# Diagnostics

The diagnostic system grew when there was nothing real to surface, so the
default became "log everything we can think of". That defaulted us
into a wide lifecycle surface (engine started, plan started, action
started, op checked, hook fired, etc.) carrying severity gradations
and chattiness levels to make the noise manageable.

The actual signal users want is much smaller: **is my config valid,
where are the issues, what would/did change, am I making progress, and
how did the run end.** Everything else is decoration.

This document rebuilds the diagnostic surface around just that
signal.

## What survives

1. **Diagnostics** - "your config / environment is wrong here, here's
   why, here's how to fix it". With source span, severity, hint, help.
2. **Changes** - "this op would change X" (planned) or "this op did
   change X" (executed). Drift detection output, apply-time mutation
   reports.
3. **Progress** - "currently connecting to host:foo", "currently
   checking action 3/12: posix.user(alice)". A status line, latest-
   wins on TTY; appended as one line per update on non-TTY.
4. **Command-specific outputs** - `inspect` dumps resolved values,
   `index` dumps the step catalog, `graph` dumps the cross-deploy
   DAG. These are the actual output of those commands, not
   lifecycle.
5. **Summary** - at end of run: N changed, M failed, K skipped, time.
   One line. Computed by the CLI from the final `ExecutionReport`;
   not an event type.

## What dies

The entire lifecycle surface:

| Goes away                                                                                                    | Why                                                                                  |
| ------------------------------------------------------------------------------------------------------------ | ------------------------------------------------------------------------------------ |
| `event.EngineEvent` + `EngineStarted`/`EngineConnecting`/`EngineFinished` factories                          | Nobody cares the engine started. "Connecting to X" becomes a Progress update.        |
| `event.PlanEvent` + `PlanStarted`/`StepPlanned`/`PlanProduced`/`PlanFinished` factories                      | The user ran the command; planning happened. The result of planning is what matters. |
| `event.ActionEvent` + `ActionStarted`/`ActionFinished`/`HookTriggered`/`HookSkipped` factories               | Per-action lifecycle is filler. Per-action changes/diagnostics carry the signal.     |
| `event.OpEvent` + `OpCheckStarted`/`OpChecked`/`OpExecuteStarted`/`OpExecuted` factories                     | Same reasoning, finer granularity.                                                   |
| `event.Chattiness` enum (`Subtle`/`Reserved`/`Normal`/`Chatty`/`Yappy`) + every `Chattiness` field on events | Existed to dial down lifecycle noise. With no lifecycle, no dialing needed.          |
| Four `Raise{Engine,Plan,Action,Op}Diagnostic` factories                                                      | Collapsed to one factory; scope was not a useful axis.                               |
| Four `event.{Engine,Plan,Action,Op}Diagnostic` envelope types                                                | Collapsed to one `event.Diagnostic`.                                                 |
| `policyEmitter.seen` + `mu` + `shouldEmit` + `Policy.DedupDiagnostics`                                       | Dedup is a consumer concern if it's wanted at all. Emitter is stateless.             |
| `Policy.SuppressPlan`                                                                                        | No lifecycle to suppress.                                                            |
| `event.{Engine,Plan,Action,Op}Kind` enums + their `*_string.go` files                                        | Discriminators for the deleted types.                                                |

## Target types

### Cause

Optional contextual tag. Most events have no notable trigger;
hooks are the first thing that does. Renderer decides whether to
display.

```go
type CauseKind uint8

const (
    CauseNone CauseKind = iota // zero value - no notable trigger
    CauseHook
    // grow as new triggers appear (deferred resource arrival,
    // scheduled re-eval, retry context)
)

type Cause struct {
    Kind CauseKind
    Ref string // hook ID for CauseHook; empty for CauseNone
}
```

Value type, not pointer. `Cause{}` is the "no trigger" case;
nothing has to be allocated for the default.

### Severity

Three levels, down from six:

```go
type Severity uint8

const (
    SeverityInfo Severity = iota
    SeverityWarning
    SeverityError
)
```

`SeverityInfo` is for non-actionable notices (e.g. "deferred resource
will be satisfied later"). `SeverityWarning` and `SeverityError` map
to LSP `Warning` and `Error`. `WarningsAsErrors` policy keeps working
as a pure remap. The old `Notice`/`Debug`/`Fatal` levels existed to
gate lifecycle chattiness; they go with lifecycle.

### Diagnostic

```go
type Diagnostic struct {
    Time time.Time
    Severity Severity
    Template Template // existing event.Template - ID, Text, Hint, Help, Data, Source
    Cause Cause
}
```

One type. No scope, no chattiness, no envelope-per-phase. The
`Template.Source *spec.SourceSpan` tells you where in the file the
diagnostic refers to - that's the only "where" the user needs.

### Change

```go
type Change struct {
    Time time.Time
    Phase ChangePhase // Planned (would-change) or Executed (did-change)
    Step StepRef // which step produced this change
    DisplayID string // op display ID
    Drift spec.DriftDetail // existing type - Field, Want, Got
    Cause Cause
}

type ChangePhase uint8

const (
    ChangePlanned ChangePhase = iota // check-only or pre-execute
    ChangeExecuted // post-execute
)

type StepRef struct {
    Index int
    Kind string
    Desc string
}
```

Changes are emitted live as they're discovered (during check) or as
they happen (during apply). Same shape both phases, distinguished by
`Phase`. Consumer sums them into the final summary.

### Progress

```go
type Progress struct {
    Time time.Time
    Text string // human-readable status
}
```

No `Cause` - Progress is too ephemeral to bother. No severity - 
it's not a diagnostic. Just a string and a timestamp. The CLI on TTY
replaces the current status line with `Text`; on non-TTY (CI, piped to
file), appends one line `<timestamp> <text>` to the stream.

Examples:

- `connecting to host:tikrlinux`
- `checking action 3/12: posix.user(alice)`
- `executing action 5/12: container.instance(nginx)`

## Target Emitter

```go
type Emitter interface {
    EmitDiagnostic(d Diagnostic)
    EmitChange(c Change)
    EmitProgress(p Progress)

    EmitInspect(e InspectEvent)
    EmitIndexAll(e IndexAllEvent)
    EmitIndexStep(e IndexStepEvent)
    EmitGraph(e GraphEvent)
}
```

Seven methods, down from twelve. The first three are the live event
stream during normal runs. The last four are command-specific outputs
that fire only when the corresponding command runs (`inspect`,
`index`, the graph view); they stay split because their payloads are
heterogeneous and the consumer wants strongly-typed access.

If we later notice we never need command outputs in isolation, those
four collapse into one `EmitResult(r CommandResult)` with a tagged
union. Out of scope for this rework.

## Ordering, timestamps, threading

Every event has a `Time` field set at emit. The CLI renders events
in emission order. Concurrent emits (op pool, plan workers) are
serialized at the displayer boundary - single mutex around the
output write, or a buffered channel feeding one writer goroutine.
Either works; implementation detail.

Timestamps appear in non-TTY mode so the log is reconstructible. TTY
mode hides them unless the user asks for them (probably `-v` flag).

Emission order is what we guarantee. Wall-clock order is *not*
strictly guaranteed - two ops emitting at nearly the same instant
might serialize in either order. This matches today's behavior and
is the right tradeoff (lock-free producer side).

## Hook context

Engine code wraps the emitter when entering a hook:

```go
hookEmitter := emitter.WithCause(Cause{Kind: CauseHook, Ref: hookID})
runHook(ctx, hookEmitter)
```

`WithCause` returns a wrapping emitter that stamps the given Cause on
every event it forwards (Diagnostic, Change, Progress) unless the
event already has a non-zero Cause. When the hook exits, the wrap is
discarded; subsequent emits go through the unwrapped emitter with
zero Cause.

The renderer pattern-matches on `Cause.Kind`. For `CauseHook` it
prefixes events with `[hook:<ref>]` (subject to glyph contract). For
`CauseNone` it shows nothing. Future Cause kinds (`CauseRetry`,
`CauseDeferred`, etc.) get their own rendering decision when they
land.

## LSP consumer

Unchanged from the earlier draft, simplified by the smaller surface:

The LSP server constructs a `diagnostic.Emitter` for each evaluation
pass. The linker (and the plan-phase walker, when it lands) routes
every diagnostic through that Emitter. The LSP server captures each
`Diagnostic`, converts it to `protocol.Diagnostic` via the existing
`diagnosticToLSP` logic, and publishes the batch.

The LSP ignores `EmitChange`, `EmitProgress`, and the command-output
methods - none of those are file diagnostics. With the smaller event
set this filtering is one switch statement.

Plan-phase diagnostics (cycle detection, missing producers, multiple
producers - see #326, #333, #347, #367, #368) reach the LSP because
the linker walks those checks without executing targets, and they
route through the same Emitter as authoring-phase diagnostics. This
is the "killer LSP" payoff.

## Migration

### Phase 1 - new types, new emitter, shims

- Add `event.Diagnostic`, `event.Change`, `event.Progress`,
  `event.Cause`, `event.CauseKind`.
- Add new `Severity` enum (`Info`/`Warning`/`Error`).
- Add `Emitter` with the seven methods. Provide a default
  implementation that forwards old method calls to new where there's
  a meaningful mapping (diagnostic envelopes -> Diagnostic).
- Lifecycle methods on the old interface become no-ops on the new
  emitter - call sites still compile, output goes silent.

### Phase 2 - delete lifecycle producers

- Delete every call to `EngineStarted`, `EngineConnecting`,
  `EngineFinished`, `PlanStarted`, `StepPlanned`, `PlanProduced`,
  `PlanFinished`, `ActionStarted`, `ActionFinished`, `HookTriggered`,
  `HookSkipped`, `OpCheckStarted`, `OpChecked`, `OpExecuteStarted`,
  `OpExecuted` in `engine/`, `cmd/scampi/`, and anywhere else.
- "Engine connecting" call sites switch to `EmitProgress`.
- Per-op "checked" / "executed" call sites become either
  `EmitChange` (when there's something to report) or silence (when
  the op was a no-op).
- Hook entry/exit switches to `emitter.WithCause(Cause{Hook, id})`
  for the duration of hook execution.

### Phase 3 - CLI render rewrite

- Status-line renderer that handles `EmitProgress` (TTY: overwrite
  line; non-TTY: append timestamp + text).
- Inline renderer that streams `EmitDiagnostic` and `EmitChange` in
  emission order.
- Final summary computed from `ExecutionReport`, printed once at end.
- Delete the chattiness-gated render paths.

### Phase 4 - LSP wired to new Emitter

- New `linker.AnalyzeWithEmitter(ctx, path, src, emitter, ...)` entry
  point.
- `lsp.diagnosticCollector` implements `Emitter`, captures
  diagnostics, ignores everything else.
- `lsp/eval.go:evaluate` switches to the collector path.
- Plan-phase diagnostic walk (no execution) reuses engine
  graph/cycle code in check-only mode.

### Phase 5 - bare-error migration

#385 and #386 absorbed here:

1. `lang/` pipeline errors -> typed `Diagnostic` producers.
2. `mod/`, `linker/` remaining bare errors -> typed.
3. `step/`, `target/`, `secret/`, `osutil/` -> typed.

`TestBareErrorBan` (`test/rules_test.go`) gets re-enforced.

### Phase 6 - delete the old surface (point of no return)

After all phases 1-5 land:

- Delete `event.EngineEvent`, `event.PlanEvent`, `event.ActionEvent`,
  `event.OpEvent` and their kind enums + stringer files.
- Delete `event.{Engine,Plan,Action,Op}Diagnostic`.
- Delete every factory function for the above
  (`diagnostic/diagnostic.go:119-676`, most of the file).
- Delete `Raise{Engine,Plan,Action,Op}Diagnostic`.
- Delete `event.Chattiness` + its stringer.
- Delete the old `Emitter` methods (`EmitEngineLifecycle`,
  `EmitPlanLifecycle`, `EmitActionLifecycle`, `EmitOpLifecycle`,
  `EmitEngineDiagnostic`, `EmitPlanDiagnostic`,
  `EmitActionDiagnostic`, `EmitOpDiagnostic`).
- Delete `policyEmitter.seen`, `mu`, `shouldEmit`.
- Delete `Policy.DedupDiagnostics`, `Policy.SuppressPlan`.

Each phase is independently revert-able. Phase 6 is the
no-going-back boundary; do it only after every producer and consumer
is migrated.

## What gets deleted (file-level summary)

The diagnostic package shrinks dramatically:

- `diagnostic/diagnostic.go` - currently 677 lines, dominated by
  lifecycle factories (lines 119-582) and four `Raise*Diagnostic`
  helpers (lines 617-676). Post-rewrite: well under 200 lines.
- `diagnostic/policy.go` - currently 149 lines. Post-rewrite: severity
  remap only, probably under 30 lines.
- `diagnostic/event/event.go` - currently 240 lines. Post-rewrite:
  the survivor types (`Diagnostic`, `Change`, `Progress`,
  `InspectEvent`, `IndexAllEvent`, `IndexStepEvent`, `GraphEvent`,
  their supporting structs). Plus `Cause`, `Severity`, `ChangePhase`.
- `diagnostic/event/{enginekind,plankind,actionkind,opkind,chattiness}_string.go`
 - deleted.

## Open questions

1. **Verbosity flags (`-v`, `-vv`, `-vvv`).** With chattiness gone,
   what do these gate?
 - `-v`: include `Info`-severity diagnostics + timestamps on TTY?
 - `-vv`: include planned-no-change Changes (the "would not change
     this either" detail) + per-op timing?
 - `-vvv`: maybe nothing - just remove the level?
   
   Decide during phase 3 once the renderer skeleton is in.

2. **Plan-phase diagnostic walker.** The engine already does cycle
   detection, multiple-producer detection, etc. before any op runs.
   The work is to extract a `linker.PlanCheck(ctx, ...)` (or
   `engine.PlanCheck`) entry point that consumes an Emitter and runs
   those checks without touching targets. Likely small - those
   checks already don't dial out - but needs verification during
   phase 4.

3. **CLI render's status-line implementation.** A few sensible
   choices (raw ANSI cursor codes, `tea`/`bubbles`, hand-rolled
   line eraser). Pick during phase 3. Doesn't affect the contract.

## Acceptance

- One `Diagnostic` type, one `Raise` factory.
- `Emitter` interface has seven methods, all non-redundant.
- `policyEmitter` is stateless (severity remap only).
- `TestBareErrorBan` re-enabled, no bare errors in package code.
- LSP consumes from `linker.AnalyzeWithEmitter`; sees every authoring
  *and* plan-phase diagnostic the CLI sees.
- CLI status line shows current activity on TTY, streams timestamped
  lines on non-TTY.
- All lifecycle event types and their factories are gone from the
  tree.
- Total LoC in `diagnostic/` shrinks meaningfully (~60% smaller is a
  reasonable target; final number depends on phase 6's exact cut).
