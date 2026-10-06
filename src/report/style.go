package report

import "os"

// ANSI styling. Only apply to whole lines or to the last cell of a tabwriter row:
// tabwriter counts escape bytes as width, so styling a padded cell breaks alignment.
const (
	ansiReset  = "\033[0m"
	ansiBold   = "\033[1m"
	ansiDim    = "\033[2m"
	ansiYellow = "\033[1;33m"
)

// useColor is true when stdout is a terminal and the user hasn't opted out via NO_COLOR.
var useColor = detectColor()

func detectColor() bool {
	if os.Getenv("NO_COLOR") != "" || os.Getenv("TERM") == "dumb" {
		return false
	}
	info, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return info.Mode()&os.ModeCharDevice != 0
}

func style(code, s string) string {
	if !useColor {
		return s
	}
	return code + s + ansiReset
}

// Bold is used for section titles.
func Bold(s string) string { return style(ansiBold, s) }

// Dim is used for secondary information like footers and warnings.
func Dim(s string) string { return style(ansiDim, s) }

// Warn is used for alert names.
func Warn(s string) string { return style(ansiYellow, s) }
