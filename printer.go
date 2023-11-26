package rye

import (
	"fmt"

	"github.com/logrusorgru/aurora/v4"
)

var PrintVerbosly bool = false

func PrintlnInfo(format string, args ...interface{}) {
	fmt.Println(aurora.Green("•"), fmt.Sprintf(format, args...))
}

func PrintlnVerbose(format string, args ...interface{}) {
	if PrintVerbosly {
		fmt.Println(aurora.Cyan("•"), fmt.Sprintf(format, args...))
	}
}

func PrintlnError(format string, args ...interface{}) {
	fmt.Println(aurora.Red("•"), fmt.Sprintf(format, args...))
}

func SprintfInfo(format string, args ...interface{}) string {
	return aurora.Sprintf(aurora.Green("•"), fmt.Sprintf(format, args...))
}

func SprintfVerbose(format string, args ...interface{}) string {
	return aurora.Sprintf(aurora.Cyan("•"), fmt.Sprintf(format, args...))
}

func SprintfError(format string, args ...interface{}) string {
	return aurora.Sprintf(aurora.Red("•"), fmt.Sprintf(format, args...))
}
