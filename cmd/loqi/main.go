package main

import (
	"os"
	"strings"

	"github.com/danterolle/loqi/cmd/loqi/commands"
	"github.com/danterolle/loqi/config"
)

func main() {
	os.Exit(run())
}

func run() int {
	cfgPath, args := extractConfig()
	cfg, err := config.Load(cfgPath)
	if err != nil {
		commands.Fatal(err)
		return 1
	}

	if err := commands.Run(cfg, args); err != nil {
		commands.Fatal(err)
		return 1
	}
	return 0
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
