package commands

import (
	"context"
	"fmt"
	"os"

	"github.com/danterolle/voca/config"
	"github.com/danterolle/voca/translate"
)

func RunBatch(cfg *config.Config, args []string) error {
	model, from, to, fs, h, help := parseTranslateFlags("batch", args, cfg.Backend.Model)

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

	input, err := ReadStdinOrFile(fs.Args())
	if err != nil {
		fmt.Fprintf(os.Stderr, "  ✖ Error: %v\n", err)
		fmt.Fprintf(os.Stderr, "Usage: voca batch --from <lang> --to <lang> [file]\n")
		fs.PrintDefaults()
		return fmt.Errorf("no input: %w", err)
	}

	core, cleanup, err := setupRun(cfg, model)
	if err != nil {
		return err
	}
	defer cleanup()
	ctx := context.Background()

	output, err := translate.Batch(ctx, core, input, from, to)
	if err != nil {
		return err
	}

	os.Stdout.Write(output)
	fmt.Println()
	return nil
}
