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
	ANSI       string
	CLIOptions struct {
		ColorMode ColorMode
	}
	cli struct {
		opts  CLIOptions
		out   io.Writer
		err   io.Writer
		isTTY bool
	}
)

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
	// Always Info / Normal
	col := ColorsForSeverity(signal.Info)
	c.outln(col.Normal, "[engine] starting")
}

func (c *cli) EngineFinish(_ signal.Severity, nChanged, nUnits int, duration time.Duration) {
	units := pluralS(nUnits, "unit")

	if nChanged > 0 {
		// Important + Highlight
		col := ColorsForSeverity(signal.Important)
		c.outln(
			col.Highlight,
			"[engine] finished (%s, %d %s, %d %s)",
			duration,
			nChanged,
			pluralS(nChanged, "change"),
			nUnits,
			units,
		)
	} else {
		// Info + Dimmed
		col := ColorsForSeverity(signal.Info)
		c.outln(
			col.Dimmed,
			"[engine] finished (%s, no changes, %d %s)",
			duration,
			nUnits,
			units,
		)
	}
}

// Config / planning phase
// =============================================

func (c *cli) PlanStart(_ signal.Severity) {
	col := ColorsForSeverity(signal.Info)
	c.outln(col.Normal, "[plan] start")
}

func (c *cli) UnitPlanned(_ signal.Severity, index int, name, kind string) {
	col := ColorsForSeverity(signal.Debug)
	c.outln(col.Dimmed, "  [unit] #%d %s (%s)", index, name, kind)
}

func (c *cli) PlanFinish(_ signal.Severity, unitCount int, duration time.Duration) {
	col := ColorsForSeverity(signal.Info)
	c.outln(col.Dimmed, "[plan] %d %s (%s)", unitCount, pluralS(unitCount, "unit"), duration)
}

// Action lifecycle
// =============================================

func (c *cli) ActionStart(_ signal.Severity, name string) {
	// Always Notice / Normal
	col := ColorsForSeverity(signal.Notice)
	c.outln(col.Normal, "[action] %s", name)
}

func (c *cli) ActionFinish(_ signal.Severity, name string, changed bool, duration time.Duration) {
	if changed {
		// Important + Highlight
		col := ColorsForSeverity(signal.Important)
		c.outln(col.Highlight, "[action] %s changed (%s)", name, duration)
	} else {
		// Info + Dimmed (never green)
		col := ColorsForSeverity(signal.Info)
		c.outln(col.Dimmed, "[action] %s up-to-date", name)
	}
}

func (c *cli) ActionError(_ signal.Severity, name string, err error) {
	col := ColorsForSeverity(signal.Error)
	c.errln(col.Highlight, "[action] %s failed: %v", name, err)
}

// Ops signals
// =============================================

func (c *cli) OpCheckStart(_ signal.Severity, action, op string) {
	col := ColorsForSeverity(signal.Debug)
	c.outln(col.Dimmed, "  [check] %s/%s", action, op)
}

func (c *cli) OpCheckSatisfied(_ signal.Severity, action, op string) {
	col := ColorsForSeverity(signal.Debug)
	c.outln(col.Dimmed, "  [check] %s/%s ✓", action, op)
}

func (c *cli) OpCheckUnsatisfied(_ signal.Severity, action, op string) {
	// Notice / Normal (never Highlight)
	col := ColorsForSeverity(signal.Notice)
	c.outln(col.Normal, "  [check] %s/%s needs change", action, op)
}

func (c *cli) OpCheckUnknown(_ signal.Severity, action, op string, err error) {
	col := ColorsForSeverity(signal.Warning)
	c.errln(col.Normal, "  [check] %s/%s unknown: %v", action, op, err)
}

func (c *cli) OpExecuteStart(_ signal.Severity, action, op string) {
	col := ColorsForSeverity(signal.Debug)
	c.outln(col.Dimmed, "  [exec] %s/%s", action, op)
}

func (c *cli) OpExecuteFinish(_ signal.Severity, action, op string, changed bool, d time.Duration) {
	if changed {
		// Info / Normal (not Important!)
		col := ColorsForSeverity(signal.Info)
		c.outln(col.Normal, "  [exec] %s/%s changed (%s)", action, op, d)
	} else {
		col := ColorsForSeverity(signal.Debug)
		c.outln(col.Dimmed, "  [exec] %s/%s no-op", action, op)
	}
}

func (c *cli) OpExecuteError(_ signal.Severity, action, op string, err error) {
	col := ColorsForSeverity(signal.Error)
	c.errln(col.Highlight, "  [exec] %s/%s failed: %v", action, op, err)
}

// User-visible errors (expected, actionable)
// =============================================

func (c *cli) UserError(_ signal.Severity, message string) {
	col := ColorsForSeverity(signal.Error)
	c.errln(col.Normal, "[error] %s", message)
}

// Internal errors (bugs, invariants violated)
// =============================================

func (c *cli) InternalError(_ signal.Severity, message string, err error) {
	col := ColorsForSeverity(signal.Fatal)
	if err != nil {
		c.errln(col.Highlight, "[fatal] %s: %v", message, err)
	} else {
		c.errln(col.Highlight, "[fatal] %s", message)
	}
}

// Internal helpers
// =============================================

func (c *cli) outln(color Color, format string, args ...any) {
	c.println(c.out, string(colToANSI(color)), format, args...)
}

func (c *cli) errln(color Color, format string, args ...any) {
	c.println(c.err, string(colToANSI(color)), format, args...)
}

func (c *cli) println(w io.Writer, color string, format string, args ...any) {
	msg := c.paint(color, format, args...)
	_, _ = fmt.Fprintln(w, msg)
}

func (c *cli) paint(color string, format string, args ...any) string {
	if !c.shouldUseColor() {
		return fmt.Sprintf(format, args...)
	}
	return color + fmt.Sprintf(format, args...) + string(Reset)
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
