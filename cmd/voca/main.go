package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/danterolle/voca/config"
)

var Version string

func main() {
	cfgPath := extractConfig()
	cfg, err := config.Load(cfgPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "  ✖ Error: %v\n", err)
		os.Exit(1)
	}

	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "translate":
			runTranslate(cfg, os.Args[2:])
			return
		case "batch":
			runBatch(cfg, os.Args[2:])
			return
		case "-h", "--help":
			printUsage()
			return
		}
	}
	runTUI(cfg)
}

func extractConfig() string {
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
	os.Args = filtered
	return cfgPath
}

func printUsage() {
	printBanner()
	fmt.Println("Usage:")
	fmt.Println("  voca                              Start the terminal UI (default)")
	fmt.Println("  voca translate [flags] <text|file>              One-shot translation")
	fmt.Println("  voca batch [flags] <file|stdin>                 Batch translate JSON or text")
	fmt.Println()
	fmt.Println("Global flags:")
	fmt.Println("  --config <path>                   Path to config file (optional)")
	fmt.Println("  -h, --help                        Show this help message")
	fmt.Println()
	cfg := config.Default()
	fmt.Println("Configurable flags (translate/batch):")
	fs := flag.NewFlagSet("translate", flag.ExitOnError)
	fs.String("from", defaultFrom, "source language code")
	fs.String("to", defaultTo, "target language code")
	fs.String("model", cfg.Backend.Model, "translation model")
	fs.PrintDefaults()
}
