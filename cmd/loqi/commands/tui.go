package commands

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

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

	logDiag := func(format string, args ...any) {
		fmt.Fprintf(os.Stderr, format, args...)
	}

	core, cleanup, err := setup.SetupRun(cfg, model, logDiag, func() { printBanner(false) })
	if err != nil {
		return err
	}
	defer cleanup()

	logDiag("\n  Starting TUI...")
	time.Sleep(800 * time.Millisecond)
	logDiag("\n")

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	return tui.RunBubbleTea(ctx, core.Backend, core.Languages)
}
