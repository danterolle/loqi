package ollama

import (
	"fmt"
	"strings"
)

func renderPullStatus(status string, total, completed int64) {
	if total > 0 {
		pct := float64(completed) / float64(total) * 100
		bar := progressBar(pct, 30)
		fmt.Printf("\r     %s  %.0f%%", bar, pct)
	} else if status == "success" {
		fmt.Printf("\r     %s  100%%\n", progressBar(100, 30))
	} else if strings.Contains(status, "pulling") {
		parts := strings.SplitN(status, " ", 2)
		if len(parts) == 2 {
			short := parts[1]
			if len(short) > 12 {
				short = short[:12]
			}
			fmt.Printf("\r     Pulling %s...", short)
		}
	} else if status == "verifying sha256 digest" {
		fmt.Printf("\r     Verifying...")
	} else if status == "writing manifest" {
		fmt.Printf("\r     Writing manifest...")
	} else {
		fmt.Printf("\r     %s", status)
	}
}

func progressBar(pct float64, width int) string {
	filled := int(pct * float64(width) / 100)
	if filled > width {
		filled = width
	}
	return strings.Repeat("█", filled) + strings.Repeat("░", width-filled)
}
