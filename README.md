# VOCA

[![Go Version](https://img.shields.io/badge/Go-1.26-00ADD8?logo=go)](https://go.dev)
[![License](https://img.shields.io/badge/license-Apache%202.0-blue)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/danterolle/voca)](https://goreportcard.com/report/github.com/danterolle/voca)

Local-first translation tool for desktop and developer workflows. Translate text, files, docs, and structured content using local LLMs.

**Why VOCA?** Every translation stays on your machine. No data sent to Google, DeepL or whatever. Designed for the terminal, so you can make it scriptable, pipeable, and integrable into your existing workflow. Replaces manual copy-pasting to DeepL/Google Translate when working on a computer.

---

## Table of Contents

- [Installation](#installation)
- [Quick start](#quick-start)
- [TUI mode](#tui-mode)
- [CLI mode](#cli-mode-one-shot)
- [Benchmarks](#benchmarks)

## Installation

```bash
# Install from source
go install github.com/danterolle/voca@latest
git clone https://github.com/danterolle/voca
cd voca
make build
```

**Prerequisites:** [Ollama](https://ollama.com) with at least one model pulled (e.g. `ollama pull gemma3:1b`).

**Makefile:**
```bash
make build
make build-linux        # cross-compile for Linux
make build-windows      # cross-compile for Windows
make build-darwin       # cross-compile for macOS
make run ARGS="--model phi4-mini:latest"
make stop               # kill ollama
```

---

## Quick start

```bash
voca

# Run a one-shot translation
voca translate --from en --to it "Hello world"

# Use a different model
voca --model phi4-mini:latest
```

## TUI mode

Interactive terminal interface with auto-translate as you type.

**Flags:**
```
--model   Ollama model to use (default: gemma4:e2b-it-qat)
```

**Keyboard:**
| Key | Action |
|-----|--------|
| `Tab` / `Shift+Tab` | Cycle focus (source → target → input) |
| `←` `→` | Change language when focused |
| `Ctrl+Y` | Copy translation to clipboard |
| `Ctrl+L` | Clear input and output |
| `Ctrl+T` | Swap source/target languages |
| `Ctrl+C` / `Esc` | Quit |

## CLI mode (one-shot)

Translate directly from the command line in one shot combo:

```bash
# Translate a string
voca translate --from en --to it "Hello world"

# Pipe from stdin
echo "Hello world" | voca translate --from en --to it

# Translate a file
voca translate --from auto --to en ./document.md

# Choose a model
voca translate --model phi4-mini:latest --from fr --to en "Bonjour le monde"

# Test with literary text (see test_data/)
voca translate --from it --to en test_data/malavoglia.md
```

**Flags:**
```
--from    Source language code (default: auto)
--to      Target language code (default: en)
--model   Ollama model (default: gemma4:e2b-it-qat)
```

## Benchmarks

See [BENCHMARKS.md](BENCHMARKS.md) for multi-sentence translation quality and speed benchmarks across 6 models and 14 languages.

## License

Apache 2.0 — see [LICENSE](LICENSE).
