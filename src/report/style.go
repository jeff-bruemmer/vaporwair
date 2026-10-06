package report

import "os"

// ANSI styling. Output is black and white: bold is the only style, used for section titles.
// Only apply to whole lines or to the last cell of a tabwriter row:
// tabwriter counts escape bytes as width, so styling a padded cell breaks alignment.
const (
	ansiReset = "\033[0m"
	ansiBold  = "\033[1m"
)

// useColor is true when stdout is a terminal and the user hasn't opted out via NO_COLOR.
var useColor = detectColor()

func detectColor() bool {
	if os.Getenv("NO_COLOR") != "" || os.Getenv("TERM") == "dumb" {
		return false
	}
	return IsTerminal(os.Stdout)
}

// IsTerminal reports whether f is attached to a terminal rather than a pipe or file.
func IsTerminal(f *os.File) bool {
	info, err := f.Stat()
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
