//go:generate stringer -type=ColorMode
package signal

type ColorMode uint8

const (
	ColorAuto ColorMode = iota
	ColorAlways
	ColorNever
)
