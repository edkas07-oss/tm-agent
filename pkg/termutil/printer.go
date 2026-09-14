package termutil

import (
	"fmt"
	"os"
)

const (
	ColorReset  = "\033[0m"
	ColorRed    = "\033[31m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
	ColorBlue   = "\033[34m"
	ColorCyan   = "\033[36m"
	ColorBold   = "\033[1m"
)

// PrintInfo prints an informational message with a blue indicator.
func PrintInfo(format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	fmt.Printf("%sℹ INFO:%s %s\n", ColorBlue, ColorReset, msg)
}

// PrintSuccess prints a success message with a green indicator.
func PrintSuccess(format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	fmt.Printf("%s✔ SUCCESS:%s %s\n", ColorGreen, ColorReset, msg)
}

// PrintWarning prints a warning message with a yellow indicator.
func PrintWarning(format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	fmt.Printf("%s▲ WARNING:%s %s\n", ColorYellow, ColorReset, msg)
}

// PrintError prints an error message to stderr with a red indicator.
func PrintError(format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	fmt.Fprintf(os.Stderr, "%s✖ ERROR:%s %s\n", ColorRed, ColorReset, msg)
}
