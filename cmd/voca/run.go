package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/danterolle/voca/config"
	"github.com/danterolle/voca/translate"
	"github.com/danterolle/voca/translate/ollama"
	"github.com/danterolle/voca/tui"
)

const defaultFrom = "auto"
const defaultTo = "en"

func readFloat64Option(opts map[string]any, key string) (float64, bool) {
	v, ok := opts[key]
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

func newCore(cfg *config.Config, model string) (*translate.Core, error) {
	prompt := translate.NewDefaultPrompt()

	var backend *ollama.Backend
	switch cfg.Backend.Type {
	case "ollama":
		backend = ollama.NewBackend(cfg.Backend.BaseURL, model, prompt)
	default:
		return nil, fmt.Errorf("unsupported backend type: %q", cfg.Backend.Type)
	}

	if np, ok := readFloat64Option(cfg.Backend.Options, "num_predict"); ok {
		backend.NumPredict = int(np)
	}
	if to, ok := readFloat64Option(cfg.Backend.Options, "timeout"); ok {
		backend.Client.Timeout = time.Duration(to) * time.Second
	}
	if t, ok := readFloat64Option(cfg.Backend.Options, "temperature"); ok {
		backend.Temperature = t
	}
	if p, ok := readFloat64Option(cfg.Backend.Options, "top_p"); ok {
		backend.TopP = p
	}
	return translate.NewCore(backend, translate.NewStaticLanguages()), nil
}

func setupRun(cfg *config.Config, model string) (*translate.Core, func()) {
	printBanner()
	ollamaCmd, started := setupOllama(model)

	var cleanup func()
	if started && ollamaCmd != nil {
		c := ollamaCmd
		cleanup = func() { _ = c.Process.Kill() }
	} else {
		cleanup = func() {}
	}

	core, err := newCore(cfg, model)
	if err != nil {
		cleanup()
		fatal(err)
	}

	return core, cleanup
}

func runTranslate(cfg *config.Config, args []string) {
	model := cfg.Backend.Model
	from := defaultFrom
	to := defaultTo

	fs := flag.NewFlagSet("translate", flag.ExitOnError)
	fs.StringVar(&model, "model", model, "translation model")
	fs.StringVar(&from, "from", from, "source language code")
	fs.StringVar(&to, "to", to, "target language code")
	h := fs.Bool("h", false, "show help")
	help := fs.Bool("help", false, "show help")
	fs.Parse(args)

	if *h || *help {
		printBanner()
		fmt.Println("Usage: voca translate [flags] <text|file>")
		fmt.Println()
		fs.PrintDefaults()
		fmt.Println()
		fmt.Println("Examples:")
		fmt.Println(`  voca translate --from it --to en "Ciao mondo!"`)
		fmt.Println("  voca translate --from en --to fr < README.md")
		os.Exit(0)
	}

	text, err := readInput(fs.Args())
	if err != nil {
		fatal(err)
	}
	if text == "" {
		fmt.Fprintf(os.Stderr, "Usage: voca translate --from <lang> --to <lang> [text|file|stdin]\n")
		fs.PrintDefaults()
		os.Exit(1)
	}

	core, cleanup := setupRun(cfg, model)
	defer cleanup()
	ui := tui.NewCLIUI(from, to, text)
	if err := ui.Run(context.Background(), core); err != nil {
		fatal(err)
	}
}

func runBatch(cfg *config.Config, args []string) {
	model := cfg.Backend.Model
	from := defaultFrom
	to := defaultTo

	fs := flag.NewFlagSet("batch", flag.ExitOnError)
	fs.StringVar(&model, "model", model, "translation model")
	fs.StringVar(&from, "from", from, "source language code")
	fs.StringVar(&to, "to", to, "target language code")
	h := fs.Bool("h", false, "show help")
	help := fs.Bool("help", false, "show help")
	fs.Parse(args)

	if *h || *help {
		printBanner()
		fmt.Println("Usage: voca batch [flags] [file]")
		fmt.Println()
		fs.PrintDefaults()
		fmt.Println()
		fmt.Println("Examples:")
		fmt.Println(`  voca batch --from en --to it < locales/en.json`)
		fmt.Println(`  voca batch --from en --to it locales/en.json`)
		fmt.Println(`  voca batch --from en --to fr README.md`)
		fmt.Println(`  echo "Hello world" | voca batch --from en --to it`)
		os.Exit(0)
	}

	input, err := readStdinOrFile(fs.Args())
	if err != nil {
		fmt.Fprintf(os.Stderr, "  ✖ Error: %v\n", err)
		fmt.Fprintf(os.Stderr, "Usage: voca batch --from <lang> --to <lang> [file]\n")
		fs.PrintDefaults()
		os.Exit(1)
	}
	if len(input) == 0 {
		fmt.Fprintf(os.Stderr, "  ✖ Error: empty input\n")
		os.Exit(1)
	}

	core, cleanup := setupRun(cfg, model)
	defer cleanup()
	ctx := context.Background()

	output, err := translate.Batch(ctx, core, input, from, to)
	if err != nil {
		fatal(err)
	}

	os.Stdout.Write(output)
	fmt.Println()
}

func runTUI(cfg *config.Config, args []string) {
	model := cfg.Backend.Model
	fs := flag.NewFlagSet("tui", flag.ExitOnError)
	fs.StringVar(&model, "model", model, "translation model")
	fs.Parse(args)

	core, cleanup := setupRun(cfg, model)
	defer cleanup()

	fmt.Printf("\n  Starting terminal interface...")
	time.Sleep(800 * time.Millisecond) // let banner finish rendering before TUI takes over alternate screen
	fmt.Printf("\n")

	ui := tui.NewBubbleTeaUI()
	if err := ui.Run(context.Background(), core); err != nil {
		fmt.Fprintf(os.Stderr, "  ✖ Error: %v\n", err)
	}
}
