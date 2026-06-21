package commands

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/danterolle/loqi/config"
	"github.com/danterolle/loqi/translate"
	"github.com/danterolle/loqi/translate/argos"
	"github.com/danterolle/loqi/translate/setup"
)

func printBatchHelp(flags *translateFlags) {
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
}

func runBatchMarkdownOrCLI(ctx context.Context, core *translate.Translator, input []byte, flags *translateFlags) error {
	if detectMarkdown(flags) {
		result, err := translate.TranslateMarkdown(ctx, core, string(input), flags.From, flags.To)
		if err != nil {
			return err
		}
		fmt.Println(result)
		return nil
	}
	output, err := translate.Batch(ctx, core, input, flags.From, flags.To)
	if err != nil {
		return err
	}
	fmt.Println(string(output))
	return nil
}

func RunBatch(cfg *config.Config, args []string) error {
	flags, err := parseTranslateFlags("batch", args, cfg)
	if err != nil {
		return err
	}
	cfg.Backend.Type = flags.Backend
	if flags.Backend == "argos" && cfg.Backend.BaseURL == config.DefaultBaseURL {
		cfg.Backend.BaseURL = argos.DefaultBaseURL
	}

	if flags.Help {
		printBatchHelp(flags)
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

	core, cleanup, err := setup.SetupRun(cfg, flags.Model, func(format string, args ...any) { logDiag(flags.Quiet, format, args...) }, func() { printBanner(flags.Quiet) })
	if err != nil {
		return err
	}
	defer cleanupRun(cleanup)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	return runBatchMarkdownOrCLI(ctx, core, input, flags)
}
