package commands

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/danterolle/voca/config"
	"github.com/danterolle/voca/tui"
)

func RunTUI(cfg *config.Config, args []string) {
	model := cfg.Backend.Model
	fs := flag.NewFlagSet("tui", flag.ExitOnError)
	fs.StringVar(&model, "model", model, "translation model")
	fs.Parse(args)

	core, cleanup, err := setupRun(cfg, model)
	if err != nil {
		Fatal(err)
	}
	defer cleanup()

	fmt.Printf("\n  Starting terminal interface...")
	time.Sleep(800 * time.Millisecond) // let banner finish rendering before TUI takes over alternate screen
	fmt.Printf("\n")

	if err := tui.RunBubbleTea(context.Background(), core); err != nil {
		fmt.Fprintf(os.Stderr, "  ✖ Error: %v\n", err)
	}
}
