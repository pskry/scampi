# Strategic Anchors

This document fixes the development order. Issues and bugs accumulate
faster than they get fixed; the only work that compounds is foundational
work that reshapes how every future issue gets handled. Anything not on
this list **waits** until the active anchor lands.

## The order

1. **Diagnostics rework** - `doc/design/diagnostics.md`.
   Foundation. Tightens the emit surface, kills emitter state, closes
   the bare-error gap, lets the LSP consume authoring-time diagnostics
   coherently. Everything downstream benefits.
   Folds in: #385, #386, #251, #329 (already fixed but design carries
   over), #367.

2. **LSP as first-class diagnostic consumer** - depends on (1).
   With the new diagnostic surface, the LSP migrates from
   `linker.Analyze` return-value parsing to direct event consumption.
   Same rewrite absorbs the LSP correctness backlog instead of
   patching each issue separately.
   Folds in: #327, #328, #338, #352, #379, #407, #408, #224.

3. **Editor plugins** - `scampi.nvim` (#162) and the Emacs
   `scampi-ts-mode` package (#413).
   The user-facing channel for (2). Done only after the LSP rewrite
   stabilizes - otherwise the plugins expose a moving target.

## Hard gate before external release

Security batch: #308, #309, #311, #313, #314, #318. None of these block
strategic work because there are no external users yet. All MUST be
green before any public release. Do them as one push when the timing is
right (probably after anchor 2 lands).

## Everything else waits

The remaining v1.0 backlog (#326, #333, #347, #368, #316, #350,
#363-#366, #94) sits unmilestoned-by-action. Most of those issues will
either resolve incidentally during anchor work (because they live in
adjacent code) or stay as targeted follow-ups after anchor 3.

PVE and Newcomer hook milestones stay dormant. PVE is gated on
baseline scampi quality; Newcomer hook resumes when anchor 3 ships
something demo-able.

## Working rules

- **One anchor in flight at a time.** While an anchor is open, do not
  read `list-issues`, do not browse the backlog, do not act on review
  tier-N punch lists. File-and-forget for any new discoveries.
- **Side discoveries during anchor work are fine** if they're in the
  anchor's scope. Out-of-scope finds get a `.issues/eval-inbox/` draft
  and disappear from working memory.
- **Each anchor produces a writeup** documenting what was rebuilt and
  what got deleted, so the next anchor inherits a clear starting point.

## Out of scope for this document

This is *direction*, not a sprint plan. No dates, no estimates, no
phase budgets. Each anchor's design doc carries its own migration
plan with concrete steps.
