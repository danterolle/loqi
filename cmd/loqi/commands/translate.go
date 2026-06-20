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

func RunTranslate(cfg *config.Config, args []string) error {
	model, from, to, fs, h, help := parseTranslateFlags("translate", args, cfg.Backend.Model)

	if *h || *help {
		printBanner()
		fmt.Println("Usage: loqi translate [flags] <text|file>")
		fmt.Println()
		fs.PrintDefaults()
		fmt.Println()
		fmt.Println("Examples:")
		fmt.Println(`  loqi translate --from it --to en "Ciao mondo!"`)
		fmt.Println("  loqi translate --from en --to fr < README.md")
		return nil
	}

	if err := validateLangs(from, to); err != nil {
		return err
	}

	text, err := ReadInput(fs.Args())
	if err != nil {
		return err
	}
	if text == "" {
		fmt.Fprintf(os.Stderr, "Usage: loqi translate --from <lang> --to <lang> [text|file|stdin]\n")
		fs.PrintDefaults()
		return fmt.Errorf("no input text or file provided")
	}

	core, cleanup, err := setup.SetupRun(cfg, model, logDiag, printBanner)
	if err != nil {
		return err
	}
	defer cleanup()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	if err := RunCLI(ctx, core, from, to, text); err != nil {
		return err
	}
	return nil
}

func RunCLI(ctx context.Context, core *translate.Core, source, target, text string) error {
	result, err := core.Translate(ctx, text, source, target)
	if err != nil {
		return err
	}
	fmt.Println(result)
	return nil
}
