package commands

import (
	"context"
	"fmt"
	"os"

	"github.com/danterolle/voca/config"
	"github.com/danterolle/voca/tui"
)

func RunTranslate(cfg *config.Config, args []string) error {
	model, from, to, fs, h, help := parseTranslateFlags("translate", args, cfg.Backend.Model)

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

	text, err := ReadInput(fs.Args())
	if err != nil {
		return err
	}
	if text == "" {
		fmt.Fprintf(os.Stderr, "Usage: voca translate --from <lang> --to <lang> [text|file|stdin]\n")
		fs.PrintDefaults()
		return fmt.Errorf("no input text or file provided")
	}

	core, cleanup, err := setupRun(cfg, model)
	if err != nil {
		return err
	}
	defer cleanup()
	if err := tui.RunCLI(context.Background(), core, from, to, text); err != nil {
		return err
	}
	return nil
}
