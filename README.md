# VOCA

Local-first translation tool for desktop and developer workflows. Translate text, files, docs, and structured content using local LLMs. Replaces manual copy-pasting to DeepL / Google Translate when working on a computer.

## Features

- Everything runs on your machine — zero data leaves
- Translate as you type (600ms debounce)
- 25 source/target languages (auto-detect supported)
- Cross-platform: macOS, Windows, Linux
- Clipboard copy, language swap, clear input
- Auto-manages Ollama lifecycle (start, pull, stop)
- Drop-in model switching via `--model`

## Quick start

```bash
# Run with default model (gemma4:e2b-it-qat)
go run .

# Use a specific model
go run . --model phi4-mini:latest
```

## Usage

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

**Makefile:**
```bash
make build              # build for current OS
make build-linux        # cross-compile for Linux
make build-windows      # cross-compile for Windows
make build-darwin       # cross-compile for macOS
make run ARGS="--model phi4-mini:latest"
make stop               # kill ollama
```

## Supported languages

Auto, English, Italian, French, German, Spanish, Portuguese, Dutch, Polish, Russian, Japanese, Chinese, Korean, Arabic, Turkish, Czech, Swedish, Danish, Finnish, Greek, Romanian, Hungarian, Vietnamese, Thai, Hindi.

> The app provides 25 language labels, but actual translation quality depends on the model. Smaller models may only handle European languages well or produce nonsense on some targets. See [BENCHMARKS.md](BENCHMARKS.md) for per-model language coverage.

## Benchmarks

See [BENCHMARKS.md](BENCHMARKS.md) for multi-sentence translation quality and speed benchmarks across 6 models and 14 languages.
