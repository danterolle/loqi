package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/danterolle/loqi/cmd/loqi/commands"
	"github.com/danterolle/loqi/config"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "  ✖ Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	cfgPath, args := extractConfig()
	cfg, err := config.Load(cfgPath)
	if err != nil {
		return err
	}
	return commands.Run(cfg, args)
}

func extractConfig() (string, []string) {
	var cfgPath string
	filtered := make([]string, 0, len(os.Args))
	for i := 0; i < len(os.Args); i++ {
		if os.Args[i] == "--config" && i+1 < len(os.Args) {
			cfgPath = os.Args[i+1]
			i++
			continue
		}
		if strings.HasPrefix(os.Args[i], "--config=") {
			cfgPath = os.Args[i][len("--config="):]
			continue
		}
		filtered = append(filtered, os.Args[i])
	}
	return cfgPath, filtered
}
