package render

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/charmbracelet/x/term"
	"godoit.dev/doit/signal"
)

type (
	ANSI  string
	ANSI2 struct {
		Highlight ANSI
		Normal    ANSI
		Dimmed    ANSI
	}
	CLIOptions struct {
		ColorMode ColorMode
		Verbosity signal.Verbosity
	}
	cli struct {
		opts  CLIOptions
		out   io.Writer
		err   io.Writer
		isTTY bool
	}
)

func (a ANSI) Hl() ANSI  { return a + Bold }
func (a ANSI) Dim() ANSI { return a + Dim }

const (
	Reset ANSI = "\033[0m"

	// styles
	Bold ANSI = "\033[1m"
	Dim  ANSI = "\033[2m"

	// base colors
	Black   ANSI = "\033[30m"
	Red     ANSI = "\033[31m"
	Green   ANSI = "\033[32m"
	Yellow  ANSI = "\033[33m"
	Blue    ANSI = "\033[34m"
	Magenta ANSI = "\033[35m"
	Cyan    ANSI = "\033[36m"
	White   ANSI = "\033[37m"

	BrightBlack   ANSI = "\033[90m"
	BrightRed     ANSI = "\033[91m"
	BrightGreen   ANSI = "\033[92m"
	BrightYellow  ANSI = "\033[93m"
	BrightBlue    ANSI = "\033[94m"
	BrightMagenta ANSI = "\033[95m"
	BrightCyan    ANSI = "\033[96m"
	BrightWhite   ANSI = "\033[97m"
)

func NewCLI(opts CLIOptions) Displayer {
	return &cli{
		opts:  opts,
		out:   os.Stdout,
		err:   os.Stderr,
		isTTY: term.IsTerminal(os.Stdout.Fd()),
	}
}

// Engine lifecycle
// =============================================

func (c *cli) EngineStart(_ signal.Severity) {
	if c.v() >= signal.VV {
		c.outln(ColorEngine.Normal, "[engine] starting")
	}
}

func (c *cli) EngineFinish(_ signal.Severity, changed, ttl int, d time.Duration) {
	units := pluralS(ttl, "unit")

	switch {
	case changed > 0:
		c.outln(
			ColorChanged.Highlight,
			"[engine] finished (%s, %d %s, %d %s)",
			d, changed, pluralS(changed, "change"), ttl, units,
		)

	case c.v() > signal.Quiet:
		c.outln(
			ColorEngine.Dimmed,
			"[engine] finished (%s, no changes, %d %s)",
			d, ttl, units,
		)
	}
}

// Config / planning phase
// =============================================

func (c *cli) PlanStart(_ signal.Severity) {
	if c.v() >= signal.VV {
		c.outln(ColorPlan.Normal, "[plan] start")
	}
}

func (c *cli) UnitPlanned(_ signal.Severity, index int, name, kind string) {
	if c.v() >= signal.VVV {
		c.outln(
			ColorUpToDate.Dimmed,
			"  [unit] #%d %s (%s)",
			index, name, kind,
		)
	}
}

func (c *cli) PlanFinish(_ signal.Severity, unitCount int, d time.Duration) {
	if c.v() >= signal.VV {
		c.outln(
			ColorPlan.Dimmed,
			"[plan] %d %s (%s)",
			unitCount, pluralS(unitCount, "unit"), d,
		)
	}
}

// Action lifecycle
// =============================================

func (c *cli) ActionStart(_ signal.Severity, name string) {
	if c.v() >= signal.V {
		c.outln(ColorAction.Normal, "[action] %s", name)
	}
}

func (c *cli) ActionFinish(_ signal.Severity, name string, changed bool, d time.Duration) {
	switch {
	case changed:
		c.outln(
			ColorChanged.Highlight,
			"[action] %s changed (%s)",
			name, d,
		)

	case c.v() >= signal.V:
		c.outln(
			ColorUpToDate.Dimmed,
			"[action] %s up-to-date",
			name,
		)
	}
}

func (c *cli) ActionError(_ signal.Severity, name string, err error) {
	c.errln(
		ColorError.Highlight,
		"[action] %s failed: %v",
		name, err,
	)
}

// Ops checks
// =============================================

func (c *cli) OpCheckStart(_ signal.Severity, action, op string) {
	if c.v() >= signal.VVV {
		c.outln(ColorCheckOK.Dimmed, "  [check] %s/%s", action, op)
	}
}

func (c *cli) OpCheckSatisfied(_ signal.Severity, action, op string) {
	if c.v() >= signal.VVV {
		c.outln(ColorCheckOK.Dimmed, "  [check] %s/%s ✓", action, op)
	}
}

func (c *cli) OpCheckUnsatisfied(_ signal.Severity, action, op string) {
	if c.v() >= signal.V {
		c.outln(ColorCheckNeed.Normal, "  [check] %s/%s needs change", action, op)
	}
}

func (c *cli) OpCheckUnknown(_ signal.Severity, action, op string, err error) {
	c.errln(
		ColorWarning.Highlight,
		"  [check] %s/%s unknown: %v",
		action, op, err,
	)
}

// Ops execute
// =============================================

func (c *cli) OpExecuteStart(_ signal.Severity, action, op string) {
	if c.v() >= signal.VV {
		c.outln(ColorExecNoop.Dimmed, "  [exec] %s/%s", action, op)
	}
}

func (c *cli) OpExecuteFinish(_ signal.Severity, action, op string, changed bool, d time.Duration) {
	if changed {
		c.outln(
			ColorExecChange.Normal,
			"  [exec] %s/%s changed (%s)",
			action, op, d,
		)
	} else if c.v() >= signal.VVV {
		c.outln(
			ColorExecNoop.Dimmed,
			"  [exec] %s/%s no-op",
			action, op,
		)
	}
}

func (c *cli) OpExecuteError(_ signal.Severity, action, op string, err error) {
	c.errln(
		ColorError.Highlight,
		"  [exec] %s/%s failed: %v",
		action, op, err,
	)
}

// User-visible errors (expected, actionable)
// =============================================

func (c *cli) UserError(_ signal.Severity, message string) {
	c.errln(ColorError.Normal, "[error] %s", message)
}

// Internal errors (bugs, invariants violated)
// =============================================

func (c *cli) InternalError(_ signal.Severity, message string, err error) {
	if err != nil {
		c.errln(ColorFatal.Highlight, "[fatal] %s: %v", message, err)
	} else {
		c.errln(ColorFatal.Highlight, "[fatal] %s", message)
	}
}

// Internal helpers
// =============================================

func (c *cli) v() signal.Verbosity {
	return c.opts.Verbosity
}

func (c *cli) outln(color Color, format string, args ...any) {
	c.println(c.out, colToANSI(color), format, args...)
}

func (c *cli) outln2(color ANSI, format string, args ...any) {
	c.println(c.out, color, format, args...)
}

func (c *cli) errln(color Color, format string, args ...any) {
	c.println(c.err, colToANSI(color), format, args...)
}

func (c *cli) errln2(color ANSI, format string, args ...any) {
	c.println(c.err, color, format, args...)
}

func (c *cli) println(w io.Writer, color ANSI, format string, args ...any) {
	msg := c.paint(color, format, args...)
	_, _ = fmt.Fprintln(w, msg)
}

func (c *cli) paint(color ANSI, format string, args ...any) string {
	if !c.shouldUseColor() {
		return fmt.Sprintf(format, args...)
	}
	return string(color) + fmt.Sprintf(format, args...) + string(Reset)
}

func (c *cli) shouldUseColor() bool {
	switch c.opts.ColorMode {
	case ColorAlways:
		return true
	case ColorNever:
		return false
	case ColorAuto:
		return c.isTTY
	default:
		return false
	}
}

func colToANSI(c Color) ANSI {
	switch c {

	case DebugHighlight:
		return BrightBlack + Bold
	case DebugNormal:
		return BrightBlack
	case DebugDimmed:
		return BrightBlack + Dim

	case InfoHighlight:
		return Blue + Bold
	case InfoNormal:
		return Blue
	case InfoDimmed:
		return Blue + Dim

	case NoticeHighlight:
		return Cyan + Bold
	case NoticeNormal:
		return Cyan
	case NoticeDimmed:
		return Cyan + Dim

	case ImportantHighlight:
		return Green + Bold
	case ImportantNormal:
		return Green
	case ImportantDimmed:
		return Green + Dim

	case WarningHighlight:
		return Yellow + Bold
	case WarningNormal:
		return Yellow
	case WarningDimmed:
		return Yellow + Dim

	case ErrorHighlight:
		return Red + Bold
	case ErrorNormal:
		return Red
	case ErrorDimmed:
		return Red + Dim

	case FatalHighlight:
		return BrightRed + Bold
	case FatalNormal:
		return BrightRed
	case FatalDimmed:
		return BrightRed + Dim

	default:
		panic("unhandled Color")
	}
}
