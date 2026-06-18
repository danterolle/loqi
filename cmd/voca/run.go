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

func newCore(cfg *config.Config) *translate.Core {
	prompt := translate.NewDefaultPrompt()
	backend := ollama.NewBackend(cfg.Backend.BaseURL, cfg.Backend.Model, prompt)
	if np, ok := cfg.Backend.Options["num_predict"]; ok {
		switch v := np.(type) {
		case int:
			backend.NumPredict = v
		case float64:
			backend.NumPredict = int(v)
		}
	}
	return translate.NewCore(
		backend,
		prompt,
		translate.NewStaticLanguages(),
		cfg.Backend.Model,
	)
}

func runTranslate(cfg *config.Config, args []string) {
	model := cfg.Backend.Model
	from := "auto"
	to := "en"

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
		fmt.Fprintf(os.Stderr, "  ✖ Error: %v\n", err)
		os.Exit(1)
	}
	if text == "" {
		fmt.Fprintf(os.Stderr, "Usage: voca translate --from <lang> --to <lang> [text|file|stdin]\n")
		fs.PrintDefaults()
		os.Exit(1)
	}

	printBanner()
	ollamaCmd, started := setupOllama(model)
	if started && ollamaCmd != nil {
		defer ollamaCmd.Process.Kill()
	}

	core := newCore(cfg)
	ui := tui.NewCLIUI(from, to, text)
	if err := ui.Run(context.Background(), core); err != nil {
		fmt.Fprintf(os.Stderr, "  ✖ Error: %v\n", err)
		os.Exit(1)
	}
}

func runBatch(cfg *config.Config, args []string) {
	model := cfg.Backend.Model
	from := "auto"
	to := "en"

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

	ollamaCmd, started := setupOllama(model)
	if started && ollamaCmd != nil {
		defer ollamaCmd.Process.Kill()
	}
	core := newCore(cfg)
	ctx := context.Background()

	output, err := translate.Batch(ctx, core, input, from, to)
	if err != nil {
		fmt.Fprintf(os.Stderr, "  ✖ Error: %v\n", err)
		os.Exit(1)
	}

	os.Stdout.Write(output)
	fmt.Println()
}

func runTUI(cfg *config.Config) {
	model := cfg.Backend.Model
	flag.StringVar(&model, "model", model, "translation model")
	flag.Parse()

	printBanner()
	ollamaCmd, started := setupOllama(model)
	if started && ollamaCmd != nil {
		defer ollamaCmd.Process.Kill()
	}

	fmt.Printf("\n  Starting terminal interface...")
	time.Sleep(800 * time.Millisecond)
	fmt.Printf("\n")

	root := newCore(cfg)
	ui := tui.NewBubbleTeaUI()
	if err := ui.Run(context.Background(), root); err != nil {
		fmt.Fprintf(os.Stderr, "  ✖ Error: %v\n", err)
	}
}
