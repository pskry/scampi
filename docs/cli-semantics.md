# CLI Semantics

`doit`’s CLI output is designed for **parallel, unordered execution**.
Colors and verbosity are not decoration — they are a **semantic contract**.

---

## Color Teller

* 🟡 **Yellow** — *Change / Mutation*
  Something modified system state.

* 🟢 **Green** — *Correctness / Stability*
  The system was already correct; no change needed.

* 🔴 **Red** — *Failure*
  An error occurred. Execution could not proceed correctly.

* 🔵 **Blue** — *Narration / Context*
  Engine, plan, and action boundaries. Structural, not outcome-based.

* ⚫ **Bright black (dim)** — *Detail / Noise*
  Ops, checks, execution details. Visible only when you asked for them.

Unused colors (cyan, magenta, etc.) are **intentionally reserved** for future semantic meaning.

---

## Philosophy

* Color answers **what happened**
* Verbosity answers **how much detail you want**
* Formatting must never imply ordering or hierarchy
* Output must stay readable under concurrency

When color is enabled, **all output is colored**.
Uncolored text would imply neutrality or importance that does not exist.

> **Color conveys meaning. Verbosity adds explanation. The CLI never lies about execution.**

---

## Verbosity Ladder

Verbosity controls **how much explanation** you receive — never *what happened*.

* **default (quiet)**
  Outcomes only.
  Changed actions and final engine summary.

* **`-v`**
  *Why* changes happened.
  Unsatisfied checks, action headers, unchanged actions.

* **`-vv`**
  *How* changes happened.
  Execution details for changed ops, plan lifecycle.

* **`-vvv`**
  Full operational detail.
  Satisfied checks, full plan units, all ops visibility.

Increasing verbosity **never removes information** — it only adds context.

---

## Appendix: Why Not Ansible / Terraform Style Output?

Traditional tools assume:

* Linear execution
* Human-paced output
* Visual grouping implies execution order

`doit` assumes the opposite:

* Actions and ops run **in parallel**
* Ordering is **not stable or meaningful**
* Grouping or indentation would **lie about causality**

As a result, `doit` intentionally avoids:

* Animated or buffered output
* Progress bars or spinners
* Tree-style nesting that implies sequencing
* Color used as decoration

Instead, the CLI reports **facts**:

* What changed
* What was already correct
* What failed
* What context you are looking at

Nothing more — and nothing less.

> If the system is unordered and parallel, the output must be honest about that reality.
