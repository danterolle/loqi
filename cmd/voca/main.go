package main

import (
	"os"
	"strings"

	"github.com/danterolle/voca/cmd/voca/commands"
	"github.com/danterolle/voca/config"
)

func main() {
	cfgPath, args := extractConfig()
	cfg, err := config.Load(cfgPath)
	if err != nil {
		commands.Fatal(err)
	}

	commands.Run(cfg, args)
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
