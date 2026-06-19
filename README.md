# VOCA

[![Go Version](https://img.shields.io/badge/Go-1.26-00ADD8?logo=go)](https://go.dev)
[![License](https://img.shields.io/badge/license-Apache%202.0-blue)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/danterolle/voca)](https://goreportcard.com/report/github.com/danterolle/voca)

An experimental, hackable local-first translation tool for desktop and developer workflows. Translate text, files, docs and structured content using local LLMs via [Ollama](https://ollama.com) or [llama.cpp](https://github.com/ggml-org/llama.cpp).

**Why VOCA?** Every translation stays on your machine. No data sent to Google, DeepL or $whatever. Designed for desktop use via terminal, **not** for mobile. You can script it, pipe it and integrate it into your development workflow if needed.

This tool is also a way to learn: to see whether a small model (Gemma 1b/2b/4b/others) can handle some real translation work well enough to make me a little less dependent on big corporations. And maybe it'll be useful to others too, not just to people who call themselves programmers. Of course, you can use or download any template and use it solely for translation, and that would work just fine. This tool is designed specifically and solely to force the model to translate.

![VOCA logo](./voca.png)

## Features Summary

- **Local** — runs entirely on your machine, no data sent to third parties
- **Dual backend** — works with [Ollama](https://ollama.com) (auto-start, model auto-pull) or [llama.cpp](https://github.com/ggml-org/llama.cpp) (manual or auto-start)
- **Three modes** — interactive TUI, one-shot CLI, and batch (JSON/text)
- **Configurable** — model, temperature, top_p, num_predict, timeout per backend
- **Scriptable** — pipe-friendly, CLI flags override config, could fits CI workflows

## Table of Contents

- [Installation](#installation)
- [Quick start](#quick-start)
- [TUI mode](#tui-mode)
- [CLI mode](#cli-mode-one-shot)
- [Supported languages](#supported-languages)
- [Batch mode](#batch-mode)
- [Benchmarks](#benchmarks)
- [Technical documentation](#technical-documentation)
- [License](#license)

## Installation

**Prerequisites (choose one):**
- [Ollama](https://ollama.com) with a model pulled (e.g. `ollama pull gemma3:1b`) — **default backend**
- [llama.cpp](https://github.com/ggml-org/llama.cpp) `llama-server` serving a GGUF model on `http://localhost:8080`

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

## Configuration

voca loads settings from a YAML config file with the following priority (each level overrides the previous):

1. CLI flags (`--model`, `--from`, `--to`, etc.)
2. `--config <path>` flag
3. `VOCA_CONFIG` environment variable
4. `~/.config/voca/config.yaml`
5. Hardcoded defaults

**Example config file (`~/.config/voca/config.yaml`):**

```yaml
backend:
  model: gemma3:1b
  base_url: http://localhost:11434
  options:
    temperature: 0.0
    num_predict: 2048
```

All fields are optional. CLI flags always override config values:

```bash
# Uses model from config file
voca --config ./config.yaml translate --from it --to en "Ciao mondo"

# Override model via CLI flag
voca --config ./config.yaml translate --model phi4:latest --from it --to en "Ciao mondo"
```

### Backends

**Ollama** (default): auto-starts `ollama serve` and pulls models on demand:

```yaml
backend:
  type: ollama
  model: gemma3:1b
  base_url: http://localhost:11434
  options:
    temperature: 0.0
    num_predict: 2048
```

**llama.cpp**: connect to an existing `llama-server` or auto-start with `model_path`:

```yaml
backend:
  type: llamacpp
  model: gemma3:1b
  base_url: http://localhost:8080
  model_path: /path/to/model.gguf          # auto-start llama-server
  server_args: ["--ctx-size", "8192", "--ngl", "99"]  # extra flags
  options:
    temperature: 0.0
    num_predict: 2048
```

When `model_path` is set, voca starts `llama-server --model <path> --host <host> --port <port> <server_args...>` as a subprocess and kills it on exit.

See [`config/config.yaml`](config/config.yaml) for a full example with defaults.

## TUI mode

Interactive terminal interface with auto-translate as you type.

**Flags:**
```
--model       Model name (default: gemma4:e2b-it-qat)
-h, --help    Show usage
```

**Keyboard:**
| Key | Action |
|-----|--------|
| `Tab` / `Shift+Tab` | Cycle focus (source → target → input) |
| `↑` `↓` | Change language when focused |
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

# Choose a different model
voca translate --model phi4-mini:latest --from fr --to en "Bonjour le monde"

# Test with literary text (see test_data/)
voca translate --from it --to en test_data/malavoglia.md
```

**Flags:**
```
--from        Source language code (default: auto)
--to          Target language code (default: en)
--model       Model name (default: gemma4:e2b-it-qat)
-h, --help    Show usage with examples
```

Language codes are validated at startup: invalid codes or using `auto` as target produce a clear error with the list of supported codes.

## Supported languages

List all supported language codes and names:

```bash
voca languages
```

Current languages: `ar`, `cs`, `da`, `de`, `el`, `en`, `es`, `fi`, `fr`, `hi`, `hu`, `it`, `ja`, `ko`, `nl`, `pl`, `pt`, `ro`, `ru`, `sv`, `th`, `tr`, `vi`, `zh`, plus `auto` (source auto-detect).

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
--model       Model name (default: gemma4:e2b-it-qat)
-h, --help    Show usage with examples
```

Language codes are validated the same way as CLI mode — invalid input produces a clear error before any translation call.

## Benchmarks

See [BENCHMARKS.md](docs/BENCHMARKS.md) for multi-sentence translation quality and speed benchmarks across 3 models and 24 languages.

## Technical documentation

See [TECHNICAL.md](docs/TECHNICAL.md) for architecture, data flow, package details and design decisions.

## License

Apache 2.0 — see [LICENSE](LICENSE).
