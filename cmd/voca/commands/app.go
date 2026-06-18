package commands

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/danterolle/voca/config"
	"github.com/danterolle/voca/translate"
	"github.com/danterolle/voca/translate/ollama"
)

var Version string

const defaultFrom = "auto"
const defaultTo = "en"

func Run(cfg *config.Config, args []string) {
	if len(args) > 1 {
		switch args[1] {
		case "translate":
			RunTranslate(cfg, args[2:])
			return
		case "batch":
			RunBatch(cfg, args[2:])
			return
		case "-h", "--help":
			PrintUsage()
			return
		}
	}
	RunTUI(cfg, args[1:])
}

func PrintUsage() {
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

func Fatal(err error) {
	fmt.Fprintf(os.Stderr, "  ✖ Error: %v\n", err)
	os.Exit(1)
}

func parseTranslateFlags(name string, args []string, defaultModel string) (model, from, to string, fs *flag.FlagSet, h, help *bool) {
	model = defaultModel
	from = defaultFrom
	to = defaultTo

	fs = flag.NewFlagSet(name, flag.ExitOnError)
	fs.StringVar(&model, "model", model, "translation model")
	fs.StringVar(&from, "from", from, "source language code")
	fs.StringVar(&to, "to", to, "target language code")
	h = fs.Bool("h", false, "show help")
	help = fs.Bool("help", false, "show help")
	fs.Parse(args)
	return
}

func newCore(cfg *config.Config, model string) (*translate.Core, error) {
	prompt := translate.NewDefaultPrompt()

	var backend *ollama.Backend
	switch cfg.Backend.Type {
	case "ollama":
		backend = ollama.NewBackend(cfg.Backend.BaseURL, model, prompt)
	default:
		return nil, fmt.Errorf("unsupported backend type: %q", cfg.Backend.Type)
	}

	readFloat := func(key string) (float64, bool) {
		v, ok := cfg.Backend.Options[key]
		if !ok {
			return 0, false
		}
		switch n := v.(type) {
		case float64:
			return n, true
		case int:
			return float64(n), true
		}
		return 0, false
	}

	if np, ok := readFloat("num_predict"); ok {
		backend.NumPredict = int(np)
	}
	if to, ok := readFloat("timeout"); ok {
		backend.Client.Timeout = time.Duration(to) * time.Second
	}
	if t, ok := readFloat("temperature"); ok {
		backend.Temperature = t
	}
	if p, ok := readFloat("top_p"); ok {
		backend.TopP = p
	}
	return translate.NewCore(backend, translate.NewStaticLanguages()), nil
}

func setupRun(cfg *config.Config, model string) (*translate.Core, func()) {
	printBanner()
	ollamaCmd, started, err := SetupOllama(model)
	if err != nil {
		Fatal(err)
	}

	var cleanup func()
	if started && ollamaCmd != nil {
		c := ollamaCmd
		cleanup = func() { _ = c.Process.Kill() }
	} else {
		cleanup = func() {}
	}

	core, err := newCore(cfg, model)
	if err != nil {
		cleanup()
		Fatal(err)
	}

	return core, cleanup
}

func printBanner() {
	gradient := []string{
		"\033[38;5;255m",
		"\033[38;5;230m",
		"\033[38;5;229m",
		"\033[38;5;221m",
		"\033[38;5;215m",
		"\033[38;5;203m",
	}
	reset := "\033[0m"

	lines := []string{
		"  ██╗   ██╗ ██████╗  ██████╗ █████╗ ",
		"  ██║   ██║██╔═══██╗██╔════╝██╔══██╗",
		"  ██║   ██║██║   ██║██║     ███████║",
		"  ╚██╗ ██╔╝██║   ██║██║     ██╔══██║",
		"   ╚████╔╝ ╚██████╔╝╚██████╗██║  ██║",
		"    ╚═══╝   ╚═════╝  ╚═════╝╚═╝  ╚═╝",
	}

	fmt.Println()
	for i, line := range lines {
		if i < len(gradient) {
			fmt.Printf("%s%s%s\n", gradient[i], line, reset)
		} else {
			fmt.Printf("%s%s%s\n", gradient[len(gradient)-1], line, reset)
		}
	}
	ver := Version
	if ver == "" {
		ver = gitVersion()
	}
	if ver != "" {
		fmt.Printf("\033[1;38;5;203m                    %s%s\n", ver, reset)
	}
	fmt.Printf("       \033[38;5;203mVersatile Offline Communication Assistant%s\n", reset)
	fmt.Println()
}

func gitVersion() string {
	cmd := exec.Command("git", "describe", "--tags", "--always", "--dirty")
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}
