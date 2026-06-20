package commands

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/danterolle/loqi/config"
	"github.com/danterolle/loqi/translate"
	"github.com/danterolle/loqi/translate/setup"
)

func RunBatch(cfg *config.Config, args []string) error {
	model, from, to, fs, h, help := parseTranslateFlags("batch", args, cfg.Backend.Model)

	if *h || *help {
		printBanner()
		fmt.Println("Usage: loqi batch [flags] [file]")
		fmt.Println()
		fs.PrintDefaults()
		fmt.Println()
		fmt.Println("Examples:")
		fmt.Println(`  loqi batch --from en --to it < locales/en.json`)
		fmt.Println(`  loqi batch --from en --to it locales/en.json`)
		fmt.Println(`  loqi batch --from en --to fr README.md`)
		fmt.Println(`  echo "Hello world" | loqi batch --from en --to it`)
		return nil
	}

	if err := validateLangs(from, to); err != nil {
		return err
	}

	input, err := ReadStdinOrFile(fs.Args())
	if err != nil || input == nil {
		if err != nil {
			fmt.Fprintf(os.Stderr, "  ✖ Error: %v\n", err)
		}
		fmt.Fprintf(os.Stderr, "Usage: loqi batch --from <lang> --to <lang> [file]\n")
		fs.PrintDefaults()
		return fmt.Errorf("no input: specify a file or pipe data to stdin")
	}

	core, cleanup, err := setup.SetupRun(cfg, model, logDiag, printBanner)
	if err != nil {
		return err
	}
	defer cleanup()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	output, err := translate.Batch(ctx, core, input, from, to)
	if err != nil {
		return err
	}

	fmt.Println(string(output))
	return nil
}
