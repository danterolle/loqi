package commands

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/danterolle/loqi/config"
	"github.com/danterolle/loqi/translate"
)

var Version string

const defaultFrom = "auto"
const defaultTo = "en"

func Run(cfg *config.Config, args []string) error {
	if len(args) <= 1 || (len(args) > 1 && strings.HasPrefix(args[1], "-")) {
		if len(args) > 1 && (args[1] == "-h" || args[1] == "--help") {
			PrintUsage()
			return nil
		}
		return RunTUI(cfg, args[1:])
	}

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
	default:
		PrintUsage()
		return fmt.Errorf("unknown command %q", args[1])
	}
}

func PrintUsage() {
	printBanner(false)
	cfg := config.Default()

	fmt.Println("━━━ Usage ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("  loqi                                     Start the terminal UI")
	fmt.Println("  loqi translate [flags] <text|file>       One-shot translation")
	fmt.Println("  loqi batch     [flags] <file|stdin>      Batch translate JSON or text")
	fmt.Println("  loqi languages                           List supported language codes")
	fmt.Println("  loqi --help                              Show this help message")
	fmt.Println()
	fmt.Println("━━━ Global flags ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("  --config <path>    Path to config file")
	fmt.Println()
	fmt.Println("━━━ Backends ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("  ollama   (default)  — general-purpose LLM")
	fmt.Println("  llamacpp            — GGUF models via llama.cpp")
	fmt.Println("  argos               — argos-translate (pip install argostranslate)")
	fmt.Println("  Set backend.type in config or use --backend")
	fmt.Println()
	fmt.Println("━━━ Translate / Batch flags ━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("  --backend string  Backend type (default %q)\n", cfg.Backend.Type)
	fmt.Printf("  --from    string  Source language code (default %q)\n", defaultFrom)
	fmt.Printf("  --to      string  Target language code (default %q)\n", defaultTo)
	fmt.Printf("  --model   string  Translation model (default %q)\n", cfg.Backend.Model)
	fmt.Println("  --quiet           Suppress diagnostic output (banner, progress)")
	fmt.Println("  --markdown        Preserve markdown structure (headings, code fences, lists)")
	fmt.Println()
	fmt.Println("━━━ Examples ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println(`  loqi translate --from it --to en "Ciao mondo!"`)
	fmt.Println(`  loqi translate --backend argos "hello world"`)
	fmt.Println("  loqi batch --from en --to it < locales/en.json")
	fmt.Println(`  loqi --config config.yaml translate --from en --to it "Hello"`)
	fmt.Println()
	fmt.Println("  # See config.yaml for llama.cpp backend setup (type: llamacpp)")
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

type translateFlags struct {
	Backend  string
	Model    string
	From     string
	To       string
	Quiet    bool
	Markdown bool
	Help     bool
	FlagSet  *flag.FlagSet
}

func parseTranslateFlags(name string, args []string, cfg *config.Config) (*translateFlags, error) {
	flags := &translateFlags{
		Backend: cfg.Backend.Type,
		Model:   cfg.Backend.Model,
		From:    defaultFrom,
		To:      defaultTo,
	}

	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.StringVar(&flags.Backend, "backend", flags.Backend, "backend type (ollama, llamacpp, argos)")
	fs.StringVar(&flags.Model, "model", flags.Model, "translation model")
	fs.StringVar(&flags.From, "from", flags.From, "source language code")
	fs.StringVar(&flags.To, "to", flags.To, "target language code")
	fs.BoolVar(&flags.Quiet, "quiet", false, "suppress diagnostic output")
	fs.BoolVar(&flags.Markdown, "markdown", false, "preserve markdown structure during translation")
	h := fs.Bool("h", false, "show help")
	help := fs.Bool("help", false, "show help")

	if err := fs.Parse(args); err != nil {
		return nil, err
	}

	flags.Help = *h || *help
	flags.FlagSet = fs
	return flags, nil
}
