package commands

import (
	"flag"
	"fmt"
	"os"

	"github.com/danterolle/voca/config"
)

var Version string

const defaultFrom = "auto"
const defaultTo = "en"

func Run(cfg *config.Config, args []string) {
	if len(args) > 1 {
		switch args[1] {
		case "translate":
			if err := RunTranslate(cfg, args[2:]); err != nil {
				Fatal(err)
			}
			return
		case "batch":
			if err := RunBatch(cfg, args[2:]); err != nil {
				Fatal(err)
			}
			return
		case "-h", "--help":
			PrintUsage()
			return
		}
	}
	RunTUI(cfg, args[1:])
}

func PrintUsage() {
	printBanner()
	cfg := config.Default()

	fmt.Println("━━━ Usage ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("  voca                                     Start the terminal UI")
	fmt.Println("  voca translate [flags] <text|file>       One-shot translation")
	fmt.Println("  voca batch     [flags] <file|stdin>      Batch translate JSON or text")
	fmt.Println("  voca --help                              Show this help message")
	fmt.Println()
	fmt.Println("━━━ Global flags ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("  --config <path>    Path to config file")
	fmt.Println()
	fmt.Println("━━━ Translate / Batch flags ━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("  --from  string    Source language code (default %q)\n", defaultFrom)
	fmt.Printf("  --to    string    Target language code (default %q)\n", defaultTo)
	fmt.Printf("  --model string    Translation model (default %q)\n", cfg.Backend.Model)
	fmt.Println()
	fmt.Println("━━━ Examples ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println(`  voca translate --from it --to en "Ciao mondo!"`)
	fmt.Println("  voca batch --from en --to it < locales/en.json")
}

func Fatal(err error) {
	fmt.Fprintf(os.Stderr, "  ✖ Error: %v\n", err)
	os.Exit(1)
}

func parseTranslateFlags(name string, args []string, defaultModel string) (model, from, to string, fs *flag.FlagSet, h, help *bool) {
	model = defaultModel
	from = defaultFrom
	to = defaultTo

	fs = flag.NewFlagSet(name, flag.ExitOnError)
	fs.StringVar(&model, "model", model, "translation model")
	fs.StringVar(&from, "from", from, "source language code")
	fs.StringVar(&to, "to", to, "target language code")
	h = fs.Bool("h", false, "show help")
	help = fs.Bool("help", false, "show help")
	fs.Parse(args)
	return
}
