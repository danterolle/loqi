package commands

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/danterolle/voca/config"
	"github.com/danterolle/voca/translate"
	"github.com/danterolle/voca/translate/ollama"
)

var Version string

const defaultFrom = "auto"
const defaultTo = "en"

func Run(cfg *config.Config, args []string) {
	var err error
	if len(args) > 1 {
		switch args[1] {
		case "translate":
			err = RunTranslate(cfg, args[2:])
		case "batch":
			err = RunBatch(cfg, args[2:])
		case "-h", "--help":
			PrintUsage()
			return
		default:
			RunTUI(cfg, args[1:])
			return
		}
	} else {
		RunTUI(cfg, args[1:])
		return
	}
	if err != nil {
		Fatal(err)
	}
}

func PrintUsage() {
	printBanner()
	fmt.Println("Usage:")
	fmt.Println("  voca                              Start the terminal UI (default)")
	fmt.Println("  voca translate [flags] <text|file>              One-shot translation")
	fmt.Println("  voca batch [flags] <file|stdin>                 Batch translate JSON or text")
	fmt.Println()
	fmt.Println("Global flags:")
	fmt.Println("  --config <path>                   Path to config file (optional)")
	fmt.Println("  -h, --help                        Show this help message")
	fmt.Println()
	cfg := config.Default()
	fmt.Println("Configurable flags (translate/batch):")
	fs := flag.NewFlagSet("translate", flag.ExitOnError)
	fs.String("from", defaultFrom, "source language code")
	fs.String("to", defaultTo, "target language code")
	fs.String("model", cfg.Backend.Model, "translation model")
	fs.PrintDefaults()
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

func setupRun(cfg *config.Config, model string) (*translate.Core, func(), error) {
	printBanner()
	ollamaCmd, started, err := SetupOllama(model)
	if err != nil {
		return nil, nil, err
	}

	var cleanup func()
	if started && ollamaCmd != nil {
		c := ollamaCmd
		cleanup = func() { _ = c.Process.Kill() }
	} else {
		cleanup = func() {}
	}

	prompt := translate.NewDefaultPrompt()
	var backend *ollama.Backend
	switch cfg.Backend.Type {
	case "ollama":
		backend = ollama.NewBackend(cfg.Backend.BaseURL, model, prompt)
	default:
		cleanup()
		return nil, nil, fmt.Errorf("unsupported backend type: %q", cfg.Backend.Type)
	}

	if np, ok := readFloatOption(cfg.Backend.Options, "num_predict"); ok {
		backend.NumPredict = int(np)
	}
	if to, ok := readFloatOption(cfg.Backend.Options, "timeout"); ok {
		backend.Client.Timeout = time.Duration(to) * time.Second
	}
	if t, ok := readFloatOption(cfg.Backend.Options, "temperature"); ok {
		backend.Temperature = t
	}
	if p, ok := readFloatOption(cfg.Backend.Options, "top_p"); ok {
		backend.TopP = p
	}

	return translate.NewCore(backend, translate.NewStaticLanguages()), cleanup, nil
}

func readFloatOption(options map[string]any, key string) (float64, bool) {
	v, ok := options[key]
	if !ok {
		return 0, false
	}
	switch n := v.(type) {
	case float64:
		return n, true
	case int:
		return float64(n), true
	}
	return 0, false
}

func printBanner() {
	gradient := []string{
		"\033[38;5;255m",
		"\033[38;5;230m",
		"\033[38;5;229m",
		"\033[38;5;221m",
		"\033[38;5;215m",
		"\033[38;5;203m",
	}
	reset := "\033[0m"

	lines := []string{
		"  ██╗   ██╗ ██████╗  ██████╗ █████╗ ",
		"  ██║   ██║██╔═══██╗██╔════╝██╔══██╗",
		"  ██║   ██║██║   ██║██║     ███████║",
		"  ╚██╗ ██╔╝██║   ██║██║     ██╔══██║",
		"   ╚████╔╝ ╚██████╔╝╚██████╗██║  ██║",
		"    ╚═══╝   ╚═════╝  ╚═════╝╚═╝  ╚═╝",
	}

	fmt.Println()
	for i, line := range lines {
		if i < len(gradient) {
			fmt.Printf("%s%s%s\n", gradient[i], line, reset)
		} else {
			fmt.Printf("%s%s%s\n", gradient[len(gradient)-1], line, reset)
		}
	}
	if Version != "" {
		fmt.Printf("\033[1;38;5;203m                    %s%s\n", Version, reset)
	}
	fmt.Printf("       \033[38;5;203mVersatile Offline Communication Assistant%s\n", reset)
	fmt.Println()
}
