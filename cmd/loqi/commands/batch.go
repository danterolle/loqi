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
	flags, err := parseTranslateFlags("batch", args, cfg.Backend.Model)
	if err != nil {
		return err
	}

	logDiag := func(format string, args ...any) {
		if !flags.Quiet {
			fmt.Fprintf(os.Stderr, format, args...)
		}
	}

	if flags.Help {
		printBanner(flags.Quiet)
		fmt.Println("Usage: loqi batch [flags] [file]")
		fmt.Println()
		flags.FlagSet.PrintDefaults()
		fmt.Println()
		fmt.Println("Examples:")
		fmt.Println(`  loqi batch --from en --to it < locales/en.json`)
		fmt.Println(`  loqi batch --from en --to it locales/en.json`)
		fmt.Println(`  loqi batch --from en --to fr README.md`)
		fmt.Println(`  echo "Hello world" | loqi batch --from en --to it`)
		return nil
	}

	if err := validateLangs(flags.From, flags.To); err != nil {
		return err
	}

	input, err := ReadStdinOrFile(flags.FlagSet.Args())
	if err != nil || input == nil {
		if err != nil {
			fmt.Fprintf(os.Stderr, "  ✖ Error: %v\n", err)
		}
		fmt.Fprintf(os.Stderr, "Usage: loqi batch --from <lang> --to <lang> [file]\n")
		flags.FlagSet.PrintDefaults()
		return fmt.Errorf("no input: specify a file or pipe data to stdin")
	}

	core, cleanup, err := setup.SetupRun(cfg, flags.Model, logDiag, func() { printBanner(flags.Quiet) })
	if err != nil {
		return err
	}
	defer cleanup()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	output, err := translate.Batch(ctx, core, input, flags.From, flags.To)
	if err != nil {
		return err
	}

	fmt.Println(string(output))
	return nil
}
