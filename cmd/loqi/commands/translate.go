package commands

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/danterolle/loqi/config"
	"github.com/danterolle/loqi/translate"
	"github.com/danterolle/loqi/translate/argos"
	"github.com/danterolle/loqi/translate/setup"
)

func logDiag(quiet bool, format string, args ...any) {
	if !quiet {
		fmt.Fprintf(os.Stderr, format, args...)
	}
}

func cleanupRun(cleanup func() error) {
	if err := cleanup(); err != nil {
		fmt.Fprintf(os.Stderr, "  ⚠ cleanup: %v\n", err)
	}
}

func printTranslateHelp(flags *translateFlags) {
	printBanner(flags.Quiet)
	fmt.Println("Usage: loqi translate [flags] <text|file>")
	fmt.Println()
	flags.FlagSet.PrintDefaults()
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println(`  loqi translate --from it --to en "Ciao mondo!"`)
	fmt.Println("  loqi translate --from en --to fr < README.md")
}

func detectMarkdown(flags *translateFlags) bool {
	if flags.Markdown {
		return true
	}
	for _, a := range flags.FlagSet.Args() {
		if strings.HasSuffix(a, ".md") {
			return true
		}
	}
	return false
}

func runTranslateMarkdownOrCLI(ctx context.Context, core *translate.Translator, text string, flags *translateFlags) error {
	if detectMarkdown(flags) {
		result, err := translate.TranslateMarkdown(ctx, core, text, flags.From, flags.To)
		if err != nil {
			return err
		}
		fmt.Println(result)
		return nil
	}
	result, err := core.Translate(ctx, text, flags.From, flags.To)
	if err != nil {
		return err
	}
	fmt.Println(result)
	return nil
}

func RunTranslate(cfg *config.Config, args []string) error {
	flags, err := parseTranslateFlags("translate", args, cfg)
	if err != nil {
		return err
	}
	cfg.Backend.Type = flags.Backend
	if flags.Backend == "argos" && cfg.Backend.BaseURL == config.DefaultBaseURL {
		cfg.Backend.BaseURL = argos.DefaultBaseURL
	}

	if flags.Help {
		printTranslateHelp(flags)
		return nil
	}

	if err := validateLangs(flags.From, flags.To); err != nil {
		return err
	}

	text, err := ReadInput(flags.FlagSet.Args())
	if err != nil {
		return err
	}
	if text == "" {
		fmt.Fprintf(os.Stderr, "Usage: loqi translate --from <lang> --to <lang> [text|file|stdin]\n")
		flags.FlagSet.PrintDefaults()
		return fmt.Errorf("no input text or file provided")
	}

	core, cleanup, err := setup.SetupRun(cfg, flags.Model, func(format string, args ...any) { logDiag(flags.Quiet, format, args...) }, func() { printBanner(flags.Quiet) })
	if err != nil {
		return err
	}
	defer cleanupRun(cleanup)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	return runTranslateMarkdownOrCLI(ctx, core, text, flags)
}
