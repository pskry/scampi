package signal

type Verbosity uint8

const (
	Quiet Verbosity = iota // default
	V
	VV
	VVV
)
