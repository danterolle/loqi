# VOCA

[![Go Version](https://img.shields.io/badge/Go-1.26-00ADD8?logo=go)](https://go.dev)
[![License](https://img.shields.io/badge/license-Apache%202.0-blue)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/danterolle/voca)](https://goreportcard.com/report/github.com/danterolle/voca)

Local-first translation tool for desktop and developer workflows. Translate text, files, docs, and structured content using local LLMs.

**Why VOCA?** Every translation stays on your machine. No data sent to Google, DeepL or $whatever. Designed for desktop and terminal, **not** for mobile. You can script it, pipe it and integrate it into your development workflow. Replaces manual copy-pasting to DeepL/Google Translate when working on text, documentation, code or sensitive documents.

---

## Installation

**Prerequisites:** [Ollama](https://ollama.com) with at least one model pulled (e.g. `ollama pull gemma3:1b`).

```bash
go install github.com/danterolle/voca@latest
```

Or build from source:

```bash
git clone https://github.com/danterolle/voca && cd voca
make build
make run ARGS="--model phi4-mini:latest"  # build + run with version tag
```

## Quick start

```bash
voca                                          # TUI mode
voca translate --from it --to en "Ciao mondo" # one-shot translation
```

## TUI mode

Interactive terminal interface with auto-translate as you type.

**Flags:**
```
--model       Ollama model (default: gemma4:e2b-it-qat)
-h, --help    Show usage
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

Translate directly from the command line:

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
--from        Source language code (default: auto)
--to          Target language code (default: en)
--model       Ollama model (default: gemma4:e2b-it-qat)
-h, --help    Show usage with examples
```

## Batch mode

Translate JSON values or text files in one pass.

```bash
# Translate all string values in a JSON file
voca batch --from en --to it locales/en.json > locales/it.json

# Translate a text file
voca batch --from en --to fr README.md

# Pipe JSON or text from stdin
echo '{"msg": "Hello"}' | voca batch --from en --to it
```

Auto-detects JSON (preserves structure, translates values) vs plain text (translates whole content).

**Flags:**
```
--from        Source language code (default: auto)
--to          Target language code (default: en)
--model       Ollama model (default: gemma4:e2b-it-qat)
-h, --help    Show usage with examples
```

## Benchmarks

See [BENCHMARKS.md](BENCHMARKS.md) for multi-sentence translation quality and speed benchmarks across 6 models and 14 languages.

## License

Apache 2.0 — see [LICENSE](LICENSE).
