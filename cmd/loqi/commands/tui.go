package commands

import (
	"context"
	"flag"
	"os/signal"
	"syscall"
	"time"

	"github.com/danterolle/loqi/config"
	"github.com/danterolle/loqi/translate/setup"
	"github.com/danterolle/loqi/tui"
)

func RunTUI(cfg *config.Config, args []string) error {
	model := cfg.Backend.Model
	fs := flag.NewFlagSet("tui", flag.ExitOnError)
	fs.StringVar(&model, "model", model, "translation model")
	fs.Parse(args)

	core, cleanup, err := setup.SetupRun(cfg, model, logDiag, printBanner)
	if err != nil {
		return err
	}
	defer cleanup()

	logDiag("\n  Starting TUI...")
	if !Quiet {
		time.Sleep(800 * time.Millisecond)
	}
	logDiag("\n")

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	return tui.RunBubbleTea(ctx, core.Backend, core.Languages)
}
