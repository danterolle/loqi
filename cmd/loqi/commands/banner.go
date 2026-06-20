package commands

import (
	"fmt"
	"os"
)

func printBanner(quiet bool) {
	if quiet {
		return
	}

	stderr := func(format string, args ...any) { fmt.Fprintf(os.Stderr, format, args...) }

	gradient := []string{
		"\033[38;5;255m",
		"\033[38;5;230m",
		"\033[38;5;229m",
		"\033[38;5;227m",
		"\033[38;5;221m",
		"\033[38;5;215m",
		"\033[38;5;209m",
		"\033[38;5;203m",
	}
	reset := "\033[0m"

	lines := []string{
		"dP                   oo",
		"88                     ",
		"88 .d8888b. .d8888b. dP",
		"88 88'  `88 88'  `88 88",
		"88 88.  .88 88.  .88 88",
		"dP `88888P' `8888P88 dP",
		"                  88   ",
		"                  dP    ",
	}

	stderr("\n")
	for i, line := range lines {
		if i < len(gradient) {
			stderr("%s%s%s\n", gradient[i], line, reset)
		} else {
			stderr("%s%s%s\n", gradient[len(gradient)-1], line, reset)
		}
	}
	stderr("\n")
	if Version != "" {
		stderr("\033[1;38;5;203m       %s%s\n", Version, reset)
	}
	stderr("   \033[38;5;203mLOcal Quiet Interpreter%s\n", reset)
	stderr("\n")
}
