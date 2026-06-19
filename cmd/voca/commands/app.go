package commands

import (
	"flag"
	"fmt"
	"os"

	"github.com/danterolle/voca/config"
	"github.com/danterolle/voca/translate"
)

var Version string
var Quiet bool

const defaultFrom = "auto"
const defaultTo = "en"

func Run(cfg *config.Config, args []string) error {
	if len(args) > 1 {
		switch args[1] {
		case "translate":
			return RunTranslate(cfg, args[2:])
		case "batch":
			return RunBatch(cfg, args[2:])
		case "languages":
			fmt.Println("Supported language codes:")
			for _, l := range translate.NewStaticLanguages().List() {
				fmt.Printf("  %-6s %s\n", l.Code, l.Name)
			}
			return nil
		case "-h", "--help":
			PrintUsage()
			return nil
		}
	}
	return RunTUI(cfg, args[1:])
}

func PrintUsage() {
	printBanner()
	cfg := config.Default()

	fmt.Println("━━━ Usage ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("  voca                                     Start the terminal UI")
	fmt.Println("  voca translate [flags] <text|file>       One-shot translation")
	fmt.Println("  voca batch     [flags] <file|stdin>      Batch translate JSON or text")
	fmt.Println("  voca languages                           List supported language codes")
	fmt.Println("  voca --help                              Show this help message")
	fmt.Println()
	fmt.Println("━━━ Global flags ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("  --config <path>    Path to config file")
	fmt.Println()
	fmt.Println("━━━ Backends ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("  Supports Ollama (default) and llama.cpp.")
	fmt.Println("  Set backend.type in config: ollama | llamacpp")
	fmt.Println()
	fmt.Println("━━━ Translate / Batch flags ━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("  --from  string    Source language code (default %q)\n", defaultFrom)
	fmt.Printf("  --to    string    Target language code (default %q)\n", defaultTo)
	fmt.Printf("  --model string    Translation model (default %q)\n", cfg.Backend.Model)
	fmt.Println("  --quiet           Suppress diagnostic output (banner, progress)")
	fmt.Println()
	fmt.Println("━━━ Examples ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println(`  voca translate --from it --to en "Ciao mondo!"`)
	fmt.Println("  voca batch --from en --to it < locales/en.json")
	fmt.Println(`  voca --config config.yaml translate --from en --to it "Hello"`)
	fmt.Println()
	fmt.Println("  # See config.yaml for llama.cpp backend setup (type: llamacpp)")
}

func logDiag(format string, args ...any) {
	if !Quiet {
		fmt.Fprintf(os.Stderr, format, args...)
	}
}

func Fatal(err error) {
	fmt.Fprintf(os.Stderr, "  ✖ Error: %v\n", err)
}

func validateLangs(from, to string) error {
	if from != "auto" && !translate.IsValidLang(from) {
		return fmt.Errorf("unsupported source language %q; %s", from, translate.ListSupported())
	}
	if !translate.IsValidLang(to) {
		return fmt.Errorf("unsupported target language %q; %s", to, translate.ListSupported())
	}
	if to == "auto" {
		return fmt.Errorf("target language cannot be %q; specify a concrete language", to)
	}
	return nil
}

func parseTranslateFlags(name string, args []string, defaultModel string) (model, from, to string, fs *flag.FlagSet, h, help *bool) {
	model = defaultModel
	from = defaultFrom
	to = defaultTo
	Quiet = false

	fs = flag.NewFlagSet(name, flag.ExitOnError)
	fs.StringVar(&model, "model", model, "translation model")
	fs.StringVar(&from, "from", from, "source language code")
	fs.StringVar(&to, "to", to, "target language code")
	fs.BoolVar(&Quiet, "quiet", false, "suppress diagnostic output")
	h = fs.Bool("h", false, "show help")
	help = fs.Bool("help", false, "show help")
	fs.Parse(args)
	return
}
