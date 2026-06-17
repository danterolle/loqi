package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/danterolle/voca/translate"
	"github.com/danterolle/voca/translate/ollama"
	"github.com/danterolle/voca/tui"
)

var Version string

func main() {
	if len(os.Args) > 1 {
		if os.Args[1] == "translate" {
			runTranslate(os.Args[2:])
			return
		}
		if os.Args[1] == "batch" {
			runBatch(os.Args[2:])
			return
		}
		if os.Args[1] == "-h" || os.Args[1] == "--help" {
			printUsage()
			return
		}
	}
	runTUI()
}

func printUsage() {
	printBanner()
	fmt.Println("Usage:")
	fmt.Println("  voca                   Start the terminal UI (default)")
	fmt.Println("  voca translate [flags] <text|file>   One-shot translation")
	fmt.Println("  voca batch [flags] <file.json>       Batch translate JSON values")
	fmt.Println()
	fmt.Println("Flags:")
	fmt.Println("  -h, --help            Show this help message")
	fmt.Println()
	fmt.Println("Translate subcommand flags:")
	fs := flag.NewFlagSet("translate", flag.ExitOnError)
	fs.String("from", "auto", "source language code")
	fs.String("to", "en", "target language code")
	fs.String("model", translate.DefaultModel, "Ollama model")
	fs.PrintDefaults()
	fmt.Println()
	fmt.Println("Batch subcommand flags:")
	bs := flag.NewFlagSet("batch", flag.ExitOnError)
	bs.String("from", "auto", "source language code")
	bs.String("to", "en", "target language code")
	bs.String("model", translate.DefaultModel, "Ollama model")
	bs.PrintDefaults()
}

func setupOllama(model string) (*exec.Cmd, bool) {
	if _, err := exec.LookPath("ollama"); err != nil {
		fmt.Fprintf(os.Stderr, "ollama not found. Install it from https://ollama.com\n")
		os.Exit(1)
	}

	started := false
	var cmd *exec.Cmd

	if !ollama.Reachable() {
		fmt.Printf("  ◆ Starting Ollama... ")
		cmd = exec.Command("ollama", "serve")
		if err := cmd.Start(); err != nil {
			fmt.Fprintf(os.Stderr, "\n  ✖ Failed to start Ollama: %v\n", err)
			os.Exit(1)
		}
		started = true
		if !ollama.WaitForReady(30) {
			fmt.Fprintf(os.Stderr, "  ✖ timeout waiting for Ollama to start\n")
			if cmd != nil {
				cmd.Process.Kill()
			}
			os.Exit(1)
		}
		fmt.Printf("online\n")
	}

	if !ollama.ModelExists(model) {
		fmt.Printf("  ◆ Pulling %s...\n", model)
		if err := ollama.PullModel(model); err != nil {
			fmt.Fprintf(os.Stderr, "  ✖ Pull failed: %v\n", err)
			if started && cmd != nil {
				cmd.Process.Kill()
			}
			os.Exit(1)
		}
		fmt.Printf("  ◆ Model ready\n")
	}

	return cmd, started
}

func newCore(model string) *translate.Core {
	return translate.NewCore(
		ollama.NewBackend("http://localhost:11434", model, translate.NewDefaultPrompt()),
		translate.NewDefaultPrompt(),
		translate.NewStaticLanguages(),
		model,
	)
}

func runTranslate(args []string) {
	fs := flag.NewFlagSet("translate", flag.ExitOnError)
	from := fs.String("from", "auto", "source language code")
	to := fs.String("to", "en", "target language code")
	model := fs.String("model", translate.DefaultModel, "Ollama model")
	h := fs.Bool("h", false, "show help")
	help := fs.Bool("help", false, "show help")
	fs.Parse(args)

	if *h || *help {
		printBanner()
		fmt.Println("Usage: voca translate [flags] <text|file>")
		fmt.Println()
		fs.PrintDefaults()
		fmt.Println()
		fmt.Println("Examples:")
		fmt.Println(`  voca translate --from it --to en "Ciao mondo!"`)
		fmt.Println("  voca translate --from en --to fr < README.md")
		os.Exit(0)
	}

	text := readInput(fs.Args())
	if text == "" {
		fmt.Fprintf(os.Stderr, "Usage: voca translate --from <lang> --to <lang> [text|file|stdin]\n")
		fs.PrintDefaults()
		os.Exit(1)
	}

	printBanner()
	ollamaCmd, started := setupOllama(*model)

	core := newCore(*model)
	ui := tui.NewCLIUI(*from, *to, text)
	if err := ui.Run(context.Background(), core); err != nil {
		fmt.Fprintf(os.Stderr, "  ✖ Error: %v\n", err)
		if started && ollamaCmd != nil {
			ollamaCmd.Process.Kill()
		}
		os.Exit(1)
	}

	if started && ollamaCmd != nil {
		ollamaCmd.Process.Kill()
	}
}

func readInput(args []string) string {
	if len(args) > 0 {
		path := args[0]
		if info, err := os.Stat(path); err == nil && !info.IsDir() {
			data, err := os.ReadFile(path)
			if err == nil {
				return strings.TrimSpace(string(data))
			}
		}
		return strings.Join(args, " ")
	}
	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) == 0 {
		data, err := io.ReadAll(os.Stdin)
		if err == nil {
			return strings.TrimSpace(string(data))
		}
	}
	return ""
}

func runBatch(args []string) {
	fs := flag.NewFlagSet("batch", flag.ExitOnError)
	from := fs.String("from", "auto", "source language code")
	to := fs.String("to", "en", "target language code")
	model := fs.String("model", translate.DefaultModel, "Ollama model")
	h := fs.Bool("h", false, "show help")
	help := fs.Bool("help", false, "show help")
	fs.Parse(args)

	if *h || *help {
		printBanner()
		fmt.Println("Usage: voca batch [flags] [file.json]")
		fmt.Println()
		fs.PrintDefaults()
		fmt.Println()
		fmt.Println("Examples:")
		fmt.Println(`  voca batch --from en --to it < locales/en.json`)
		fmt.Println(`  voca batch --from en --to it locales/en.json`)
		os.Exit(0)
	}

	input, err := readStdinOrFile(fs.Args())
	if err != nil || len(input) == 0 {
		fmt.Fprintf(os.Stderr, "Usage: voca batch --from <lang> --to <lang> [file.json]\n")
		fs.PrintDefaults()
		os.Exit(1)
	}

	ollamaCmd, started := setupOllama(*model)
	core := newCore(*model)
	ctx := context.Background()

	output, err := translate.TranslateJSON(ctx, core, input, *from, *to)
	if err != nil {
		fmt.Fprintf(os.Stderr, "  ✖ Error: %v\n", err)
		if started && ollamaCmd != nil {
			ollamaCmd.Process.Kill()
		}
		os.Exit(1)
	}

	os.Stdout.Write(output)
	fmt.Println()

	if started && ollamaCmd != nil {
		ollamaCmd.Process.Kill()
	}
}

func readStdinOrFile(args []string) ([]byte, error) {
	if len(args) > 0 {
		return os.ReadFile(args[0])
	}
	return io.ReadAll(os.Stdin)
}

func runTUI() {
	model := flag.String("model", translate.DefaultModel, "Ollama model to use for translation")
	flag.Parse()

	printBanner()
	ollamaCmd, started := setupOllama(*model)

	fmt.Printf("\n  Starting terminal interface...")
	time.Sleep(800 * time.Millisecond)
	fmt.Printf("\n")

	core := newCore(*model)
	ui := tui.NewBubbleTeaUI()
	if err := ui.Run(context.Background(), core); err != nil {
		fmt.Fprintf(os.Stderr, "  ✖ Error: %v\n", err)
	}

	if started && ollamaCmd != nil {
		ollamaCmd.Process.Kill()
	}
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
