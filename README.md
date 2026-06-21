# Loqi

[![Go Version](https://img.shields.io/badge/Go-1.26-00ADD8?logo=go)](https://go.dev)
[![License](https://img.shields.io/badge/license-Apache%202.0-blue)](LICENSE)
[![Go Report Card](https://img.shields.io/goreportcard/grade/v/github.com/danterolle/loqi?logo=go)](https://goreportcard.com/report/github.com/danterolle/loqi)

A tool for producing local translation drafts via [Ollama](https://ollama.com), [llama.cpp](https://github.com/ggml-org/llama.cpp), or [argos-translate](https://github.com/argosopentech/argos-translate). Translate text, files, docs and structured content entirely on your machine.

**Why Loqi?** 

As convenient as it is, and despite all the opt-out options and privacy policies, I generally believe it's never ideal to send your data to Google or DeepL (and yes, like everyone else, I do it too). I started this project in an attempt to make myself a bit less dependent on these great technologies.

Can a small-parameter LLM actually help me achieve this? It will never give me absolute certainty, but a traditional translation engine won't either, even though it would be much faster and more efficient.

This project is an experiment. There are several features that might make it interesting down the road, but at least for now, it meets my needs.

Please note that translation quality depends on the model you choose and small models can really make mistakes so treat the output as a draft to review, not as a guaranteed result.

Ah, of course you can use or download any model and use it solely for translation, and that would work just fine. This tool is designed specifically and solely to force the model to translate.

And perhaps expanding the model's capabilities to handle data batches and more.

Read on if you're interested.

## Features Summary

- **Local**: runs entirely on your machine, no data sent to third parties
- **Three backends**: works with [Ollama](https://ollama.com) (auto-start, model auto-pull), [llama.cpp](https://github.com/ggml-org/llama.cpp) (manual or auto-start), or [argos-translate](https://github.com/argosopentech/argos-translate)
- **Three modes**: interactive TUI, one-shot CLI, and batch (JSON/text/markdown)
- **Configurable**: model, temperature, top_p, num_predict, timeout per backend
- **Scriptable**: pipe-friendly, CLI flags override config, could fits CI workflows

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
- [Ollama](https://ollama.com) with a model pulled (e.g. `ollama pull phi4-mini`) — **default backend**
- [llama.cpp](https://github.com/ggml-org/llama.cpp) `llama-server` serving a GGUF model on `http://localhost:8080`
- [argos-translate](https://github.com/argosopentech/argos-translate) — Python 3 required, installed automatically on first use

```bash
go install github.com/danterolle/loqi@latest
```

Or build from source:

```bash
git clone https://github.com/danterolle/loqi && cd loqi
make build
make run ARGS="--model phi4-mini:latest"  # build + run with version tag
```

## Quick start

```bash
loqi                                          # TUI mode
loqi translate --from it --to en "Ciao mondo" # one-shot translation
```

## Configuration

Loqi loads settings from a YAML config file with the following priority (each level overrides the previous):

1. CLI flags (`--model`, `--from`, `--to`, etc.)
2. `--config <path>` flag
3. `LOQI_CONFIG` environment variable
4. `~/.config/loqi/config.yaml`
5. Hardcoded defaults

**Example config file (`~/.config/loqi/config.yaml`):**

```yaml
backend:
  model: phi4-mini
  base_url: http://localhost:11434
  options:
    temperature: 0.0
    num_predict: 2048
```

All fields are optional. CLI flags always override config values:

```bash
# Uses model from config file
loqi --config ./config.yaml translate --from it --to en "Ciao mondo"

# Override model via CLI flag
loqi --config ./config.yaml translate --model phi4-mini:latest --from it --to en "Ciao mondo"
```

### Backends

**Ollama** (default): auto-starts `ollama serve` and pulls models on demand:

```yaml
backend:
  type: ollama
  model: phi4-mini
  base_url: http://localhost:11434
  options:
    temperature: 0.0
    num_predict: 2048
```

**llama.cpp**: connect to an existing `llama-server` or auto-start with `model_path`:

```yaml
backend:
  type: llamacpp
  model: phi4-mini
  base_url: http://localhost:8080
  model_path: /path/to/model.gguf          # auto-start llama-server
  server_args: ["--ctx-size", "8192", "--ngl", "99"]  # extra flags
  options:
    temperature: 0.0
    num_predict: 2048
```

When `model_path` is set, loqi starts `llama-server --model <path> --host <host> --port <port> <server_args...>` as a subprocess and kills it on exit.

**Argos**: offline, rule-based translation via [argos-translate](https://github.com/argosopentech/argos-translate). Argos fills the role of a fast, non-LLM translation engine — exactly the kind of lightweight backend the project was looking for (see [Why Loqi?](#why-loqi)). It runs entirely offline with no GPU needed, and for straightforward sentences it is often faster and more deterministic than an LLM-based approach. Auto-installs `argostranslate` in a Python venv and starts a local HTTP server:

```yaml
backend:
  type: argos
  base_url: http://localhost:5000
```

When `type: argos` is set, loqi creates a Python virtual environment in `~/.cache/loqi/argos-venv`, installs `argostranslate`, and starts the bundled Python server script. Python 3 must be available on the system.

Argos has no auto-detection (`--from auto` is not supported), requires Python 3 on the system, and downloads language packages on first use which adds initial latency. Translation quality and language coverage depend on the argos-translate ecosystem rather than an LLM, so output tends to be more literal and less context-aware.

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
loqi translate --from en --to it "Hello world"

# Pipe from stdin
echo "Hello world" | loqi translate --from en --to it

# Translate a file
loqi translate --from auto --to en ./document.md

# Choose a different model
loqi translate --model phi4-mini:latest --from fr --to en "Bonjour le monde"

# Test with literary text (see test_data/)
loqi translate --from it --to en test_data/malavoglia.md
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
loqi languages
```

Current languages: `ar`, `cs`, `da`, `de`, `el`, `en`, `es`, `fi`, `fr`, `hi`, `hu`, `it`, `ja`, `ko`, `nl`, `pl`, `pt`, `ro`, `ru`, `sv`, `th`, `tr`, `vi`, `zh`, plus `auto` (source auto-detect).

## Batch mode

Translate JSON values or text files in one pass.

```bash
# Translate all string values in a JSON file
loqi batch --from en --to it locales/en.json > locales/it.json

# Translate a text file
loqi batch --from en --to fr README.md

# Pipe JSON or text from stdin
echo '{"msg": "Hello"}' | loqi batch --from en --to it
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

See [BENCHMARKS.md](docs/BENCHMARKS.md) for speed comparisons across models and a context comparison between argos and Gemma 4.

## Technical documentation

See [TECHNICAL.md](docs/TECHNICAL.md) for architecture, data flow, package details and design decisions.

## License

Apache 2.0, see [LICENSE](LICENSE).
