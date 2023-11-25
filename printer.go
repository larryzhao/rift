package rye

import (
	"fmt"

	"github.com/fatih/color"
)

var PrintVerbosly bool = false

func PrintInfo(format string, args ...interface{}) {
	fmt.Printf("%s %s.\n", color.GreenString("•"), fmt.Sprintf(format, args...))
}

func PrintVerbose(format string, args ...interface{}) {
	if PrintVerbosly {
		fmt.Printf("%s %s.\n", color.CyanString("•"), fmt.Sprintf(format, args...))
	}
}

func PrintError(format string, args ...interface{}) {
	fmt.Printf("%s %s.\n", color.RedString("•"), fmt.Sprintf(format, args...))
}
