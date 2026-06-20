package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/danterolle/loqi/config"
	"github.com/danterolle/loqi/translate"
	"github.com/danterolle/loqi/translate/setup"
)

var sentences = []string{
	"I would like to book a table for two at seven o'clock this evening, preferably near the window with a view of the garden.",
	"The conference will cover topics such as artificial intelligence, machine learning, and data privacy regulations across different countries.",
	"The museum's new exhibition features over two hundred paintings and sculptures from Renaissance artists, attracting visitors from all over the world.",
}

var targets []translate.Language

func init() {
	for _, l := range translate.NewStaticLanguages().List() {
		if l.Code != "auto" {
			targets = append(targets, l)
		}
	}
}

func main() {
	model := flag.String("model", config.DefaultModel, "Ollama model")
	flag.Parse()

	cfg := config.Default()
	cfg.Backend.Model = *model

	setupStart := time.Now()
	core, cleanup, err := setup.SetupRun(cfg, *model, logDiag, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "setup error: %v\n", err)
		os.Exit(1)
	}
	setupElapsed := time.Since(setupStart)
	defer cleanup()

	totalStart := time.Now()
	totalSentences := len(sentences)
	langTimes := make(map[string][]time.Duration)

	for si, text := range sentences {
		fmt.Fprintf(os.Stderr, "\n=== Sentence %d/%d: %q ===\n", si+1, totalSentences, ellipsis(text, 60))
		for _, tgt := range targets {
			start := time.Now()
			result, err := core.Translate(context.Background(), text, "en", tgt.Code)
			elapsed := time.Since(start)
			langTimes[tgt.Code] = append(langTimes[tgt.Code], elapsed)
			if err != nil {
				fmt.Printf("[%-4s %-10s] S%d ERROR: %v\n", tgt.Code, tgt.Name, si+1, err)
			} else {
				fmt.Printf("[%-4s %-10s] S%d %v — %q\n", tgt.Code, tgt.Name, si+1, elapsed.Round(time.Millisecond), result)
			}
		}
	}

	totalElapsed := time.Since(totalStart)
	fmt.Fprintf(os.Stderr, "\n=== Summary for %s ===\n", *model)
	fmt.Fprintf(os.Stderr, "Setup time: %v\n", setupElapsed.Round(time.Millisecond))
	fmt.Fprintf(os.Stderr, "Total translate time: %v | Avg per sentence: %v\n", totalElapsed.Round(time.Millisecond), (totalElapsed / time.Duration(totalSentences)).Round(time.Millisecond))
	fmt.Fprintf(os.Stderr, "\nLanguage averages:\n")
	for _, tgt := range targets {
		times := langTimes[tgt.Code]
		if len(times) == 0 {
			continue
		}
		var sum time.Duration
		for _, t := range times {
			sum += t
		}
		avg := sum / time.Duration(len(times))
		fmt.Fprintf(os.Stderr, "  %-4s %-10s avg %v\n", tgt.Code, tgt.Name, avg.Round(time.Millisecond))
	}
}

func logDiag(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format, args...)
}

func ellipsis(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}
