package commands

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/danterolle/loqi/config"
	"github.com/danterolle/loqi/translate/setup"
	"github.com/danterolle/loqi/tui"
)

func RunTUI(cfg *config.Config, args []string) error {
	model := cfg.Backend.Model
	fs := flag.NewFlagSet("tui", flag.ContinueOnError)
	fs.StringVar(&model, "model", model, "translation model")
	if err := fs.Parse(args); err != nil {
		return err
	}

	core, cleanup, err := setup.SetupRun(cfg, model, func(format string, args ...any) {}, func() { printBanner(false) })
	if err != nil {
		return err
	}
	defer func() {
		if err := cleanup(); err != nil {
			fmt.Fprintf(os.Stderr, "  ⚠ cleanup: %v\n", err)
		}
	}()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	return tui.RunBubbleTea(ctx, core.Backend(), core.Languages())
}
