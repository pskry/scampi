package ansi

type (
	Code string
	ANSI struct {
		Reg  Code
		Bold Code
		Dim  Code
		Ul   Code
	}
)

const (
	Reset = "\x1b[0m"

	// black
	blackReg  = "\x1b[0;30m"
	blackBold = "\x1b[1;30m"
	blackDim  = "\x1b[2;30m"
	blackUl   = "\x1b[4;30m"

	// red
	redReg  = "\x1b[0;31m"
	redBold = "\x1b[1;31m"
	redDim  = "\x1b[2;31m"
	redUl   = "\x1b[4;31m"

	// green
	greenReg  = "\x1b[0;32m"
	greenBold = "\x1b[1;32m"
	greenDim  = "\x1b[2;32m"
	greenUl   = "\x1b[4;32m"

	// yellow
	yellowReg  = "\x1b[0;33m"
	yellowBold = "\x1b[1;33m"
	yellowDim  = "\x1b[2;33m"
	yellowUl   = "\x1b[4;33m"

	// blue
	blueReg  = "\x1b[0;34m"
	blueBold = "\x1b[1;34m"
	blueDim  = "\x1b[2;34m"
	blueUl   = "\x1b[4;34m"

	// purple (magenta)
	purpleReg  = "\x1b[0;35m"
	purpleBold = "\x1b[1;35m"
	purpleDim  = "\x1b[2;35m"
	purpleUl   = "\x1b[4;35m"

	// cyan
	cyanReg  = "\x1b[0;36m"
	cyanBold = "\x1b[1;36m"
	cyanDim  = "\x1b[2;36m"
	cyanUl   = "\x1b[4;36m"

	// white
	whiteReg  = "\x1b[0;37m"
	whiteBold = "\x1b[1;37m"
	whiteDim  = "\x1b[2;37m"
	whiteUl   = "\x1b[4;37m"

	// bright black
	brightBlackReg  = "\x1b[0;90m"
	brightBlackBold = "\x1b[1;90m"
	brightBlackDim  = "\x1b[2;90m"
	brightBlackUl   = "\x1b[4;90m"

	// bright red
	brightRedReg  = "\x1b[0;91m"
	brightRedBold = "\x1b[1;91m"
	brightRedDim  = "\x1b[2;91m"
	brightRedUl   = "\x1b[4;91m"

	// bright green
	brightGreenReg  = "\x1b[0;92m"
	brightGreenBold = "\x1b[1;92m"
	brightGreenDim  = "\x1b[2;92m"
	brightGreenUl   = "\x1b[4;92m"

	// bright yellow
	brightYellowReg  = "\x1b[0;93m"
	brightYellowBold = "\x1b[1;93m"
	brightYellowDim  = "\x1b[2;93m"
	brightYellowUl   = "\x1b[4;93m"

	// bright blue
	brightBlueReg  = "\x1b[0;94m"
	brightBlueBold = "\x1b[1;94m"
	brightBlueDim  = "\x1b[2;94m"
	brightBlueUl   = "\x1b[4;94m"

	// bright purple (magenta)
	brightPurpleReg  = "\x1b[0;95m"
	brightPurpleBold = "\x1b[1;95m"
	brightPurpleDim  = "\x1b[2;95m"
	brightPurpleUl   = "\x1b[4;95m"

	// bright cyan
	brightCyanReg  = "\x1b[0;96m"
	brightCyanBold = "\x1b[1;96m"
	brightCyanDim  = "\x1b[2;96m"
	brightCyanUl   = "\x1b[4;96m"

	// bright white
	brightWhiteReg  = "\x1b[0;97m"
	brightWhiteBold = "\x1b[1;97m"
	brightWhiteDim  = "\x1b[2;97m"
	brightWhiteUl   = "\x1b[4;97m"
)

var (
	Black        = ANSI{Reg: blackReg, Bold: blackBold, Dim: blackDim, Ul: blackUl}
	Red          = ANSI{Reg: redReg, Bold: redBold, Dim: redDim, Ul: redUl}
	Green        = ANSI{Reg: greenReg, Bold: greenBold, Dim: greenDim, Ul: greenUl}
	Yellow       = ANSI{Reg: yellowReg, Bold: yellowBold, Dim: yellowDim, Ul: yellowUl}
	Blue         = ANSI{Reg: blueReg, Bold: blueBold, Dim: blueDim, Ul: blueUl}
	Purple       = ANSI{Reg: purpleReg, Bold: purpleBold, Dim: purpleDim, Ul: purpleUl}
	Cyan         = ANSI{Reg: cyanReg, Bold: cyanBold, Dim: cyanDim, Ul: cyanUl}
	White        = ANSI{Reg: whiteReg, Bold: whiteBold, Dim: whiteDim, Ul: whiteUl}
	BrightBlack  = ANSI{Reg: brightBlackReg, Bold: brightBlackBold, Dim: brightBlackDim, Ul: brightBlackUl}
	BrightRed    = ANSI{Reg: brightRedReg, Bold: brightRedBold, Dim: brightRedDim, Ul: brightRedUl}
	BrightGreen  = ANSI{Reg: brightGreenReg, Bold: brightGreenBold, Dim: brightGreenDim, Ul: brightGreenUl}
	BrightYellow = ANSI{Reg: brightYellowReg, Bold: brightYellowBold, Dim: brightYellowDim, Ul: brightYellowUl}
	BrightBlue   = ANSI{Reg: brightBlueReg, Bold: brightBlueBold, Dim: brightBlueDim, Ul: brightBlueUl}
	BrightPurple = ANSI{Reg: brightPurpleReg, Bold: brightPurpleBold, Dim: brightPurpleDim, Ul: brightPurpleUl}
	BrightCyan   = ANSI{Reg: brightCyanReg, Bold: brightCyanBold, Dim: brightCyanDim, Ul: brightCyanUl}
	BrightWhite  = ANSI{Reg: brightWhiteReg, Bold: brightWhiteBold, Dim: brightWhiteDim, Ul: brightWhiteUl}
)
