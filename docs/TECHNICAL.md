## What It Is

Loqi is a terminal-based translator that runs locally through LLMs. It supports two backends: **Ollama** (default) and **llama.cpp**. It works in three modes: an interactive TUI built with bubbletea, a `loqi translate` command for one-shot translations, and `loqi batch` for bulk-translating JSON or plain text files.

The entire codebase is Go with only five external dependencies: bubbletea, bubbles, lipgloss, yaml, atotto/clipboard.

## Package Structure

```
cmd/loqi/main.go
├─ cmd/loqi/commands/
│   ├─ app.go             ── Run, PrintUsage, flag parsing, Fatal
│   ├─ translate.go       ── RunTranslate, detectMarkdown, runTranslateMarkdownOrCLI, logDiag, cleanupRun, printTranslateHelp
│   ├─ batch.go           ── RunBatch, runBatchMarkdownOrCLI, printBatchHelp
│   ├─ tui.go             ── RunTUI
│   ├─ io.go              ── input reading (args, file, stdin)
│   └─ banner.go
├─ translate/
│   ├─ interfaces.go      ── Backend, LanguageProvider
│   ├─ core.go            ── thin Backend + LanguageProvider wrapper
│   ├─ languages.go       ── language map + sorted code
│   ├─ default_prompt.go  ── system + user prompt templates
│   ├─ factory.go         ── NewBackend, option helpers, UnloadBackend
│   ├─ batch.go           ── batch entry point (JSON dispatch)
│   ├─ json_translator.go ── recursive JSON walker + worker pool
│   ├─ markdown.go        ── line-by-line markdown translation, prefix splitting
│   ├─ mock_backend.go
│   ├─ setup/             ── backend lifecycle orchestration
│   │   ├─ setup.go       ── SetupRun, unified backend dispatch
│   │   └─ server.go      ── SetupOllama, SetupLlamaCpp, StopProcess
│   ├─ argos/              ── argos-translate backend
│   │   ├─ backend.go      ── HTTP /translate client
│   │   ├─ server.go       ── venv setup, server start, health check
│   │   └─ argos_server.py ── embedded Python HTTP server (//go:embed)
│   ├─ http/              ── shared HTTP client
│   │   └─ http.go
│   ├─ ollama/
│   │   ├─ backend.go     ── HTTP /api/chat client
│   │   ├─ lifecycle.go   ── health checks, model pull/unload
│   │   └─ progress.go    ── ANSI progress bar rendering
│   └─ llamacpp/
│       ├─ backend.go     ── OpenAI-compatible /v1/chat/completions client
│       └─ lifecycle.go   ── server check, model-ready polling
├─ tui/
│   ├─ model.go/update.go/view.go
│   ├─ commands.go        ── doTranslate, copyClipboard
│   ├─ styles.go/ui.go
├─ config/                ── YAML config loader
└─ cmd/bench/             ── multi-language benchmark
```

Domain code lives in `translate` with its interfaces; `commands` handles CLI dispatch and flag parsing; `translate/setup` owns backend lifecycle and subprocess management; `tui` owns the UI; `config` loads and merges YAML.

## Backend Selection

`translate/setup.SetupRun` dispatches based on `cfg.Backend.Type`, parameterising three variables per backend:

- **Server starter** — the function to call (`SetupOllama`, `SetupLlamaCpp`, or `SetupArgos` from `setup/server.go`)
- **Backend type string** — `"ollama"`, `"llamacpp"`, or `"argos"` passed to `translate.NewBackend`
- **Unload on close** — whether to call `UnloadBackend` during cleanup (ollama only)

Every backend returns a `*translate.Core` wrapping a struct that satisfies `translate.Backend`, plus a `func()` cleanup closure.

```go
type Backend interface {
    Translate(ctx context.Context, text, source, target string) (string, error)
}
```

## TUI Mode

When the user launches `loqi` with no arguments, `Run()` falls through to `RunTUI`, which calls `SetupRun` to initialize the backend and then passes `core.Backend` and `core.Languages` directly to `RunBubbleTea` — the TUI has no dependency on `Core` itself.

The TUI follows bubbletea's Model-View-Update pattern. Here is the flow from keystroke to rendered translation:

```
Keystroke
    │
    ▼
handleTextChange
    │
    ├── leadingDone == false?
    │       │  yes ──► leadingDone = true
    │       │           lastInput = text
    │       │           doTranslate(text) ──► backend.Translate ──► parse response
    │       │           status = "Translating..."
    │       │
    │       └── no  ──► translateSeq++
    │                    debounceMsg{seq} after 600ms
    │
    ▼ (after 600ms)
handleDebounce
    │
    ├── seq != translateSeq?  ──► discard (stale)
    ├── text == lastInput?    ──► discard (no change)
    └── ok ──► lastInput = text
               doTranslate(text)
    │
    ▼
handleTranslateResult
    │
    ├── msg.text != textarea.Value()?  ──► discard (input changed while waiting)
    └── ok ──► m.output = msg.result
               status = "Ready."
    │
    ▼
View() renders:
    headerView    ──► "loqi  From: Italian  ->  To: English"
    textarea.View ──► input area
    outputView    ──► wrapped translation
    statusView    ──► "Ready.  ctrl+y:copy  ctrl+l:clear  ..."
```

The first keystroke translates immediately (`leadingDone` gate). Every subsequent keystroke increments `translateSeq` and schedules a debounce tick. If a new keystroke arrives before the tick fires, the old tick is ignored because its sequence number no longer matches. When the result arrives, it is compared against the current textarea value: if the user changed the input while waiting, the result is thrown away. This prevents the classic race where a slow response overwrites a newer translation.

The `lastInput` field exists to solve a subtle bug: without it, the debounce handler compared `m.output` (the previous translation result) against `m.textarea.Value()` (the new input). Those are different domains — input text vs. translated text — so the comparison would miss real changes. Now it compares the current input against the last input that was actually sent for translation, which is the correct check.

## CLI Mode

`loqi translate --from it --to en "Ciao mondo"` takes a simpler path:

```
parseTranslateFlags ──► ReadInput (text, file or stdin)
                             │
                             ▼
                          setup.SetupRun(cfg, model, logDiag, printBanner)
                              │
                              ├── printBanner()
                              │
                              ├── switch cfg.Backend.Type:
                              │     ollama  ──► setup.SetupOllama()
                              │                   ├── Reachable? ──► no ──► start ollama serve
                              │                   │                      ──► WaitForReady(30s)
                              │                   ├── ModelExists? ──► no ──► PullModel
                              │                   └── return cmd handle
                              │
                              │     llamacpp ──► setup.SetupLlamaCpp()
                              │                    ├── ServerRunning? ──► yes ──► wait for model
                              │                    ├── no + model_path? ──► start llama-server
                              │                    └── return cmd handle
                              │           
                              ├── build backend with config options
                              └── return *Core + cleanup()
`                             │
                              ▼
                     signal.NotifyContext(SIGINT, SIGTERM)
                              │
                              ▼
                    RunCLI(ctx, core, from, to, text)
                              │
                              ▼
                     core.Translate ──► backend.Translate
                              │
                              ▼
                      fmt.Println(result)`
```

The signal context ensures that if the user presses CTRL+C while translating, the deferred `cleanup()` runs — which kills the subprocess only if Loqi started it. This distinction matters: if the backend was already running when Loqi launched, cleanup is a no-op.

## Batch Mode

`loqi batch --from en --to it < locales/en.json` handles JSON and plain text differently:

```
Input bytes
    │
    ├── json.Valid?
    │       │
    │       yes ──► Unmarshal into any ──► translateJSON(ctx, core, &data, from, to)
    │       │                                      │
    │       │                                      ▼
    │       │                              recursive processNode(&val)
    │       │                                      │
    │       │                         ┌────────────┼───────┐
    │       │                         ▼            ▼       ▼
    │       │                      string         map    slice
    │       │                         │            │       │
    │       │                  translateString   worker  worker
    │       │                    (max 3           pool    pool
    │       │                     concurrent)
    │       │                                  │
    │       │                                  ▼
    │       │                        json.MarshalIndent(data)
    │       │                                  │
    │       │                                  ▼
    │       │                                result
    │       │
    │       │
    │       └── no ──► core.Translate(ctx, text, from, to)
    │                                  │
    │                                  ▼
    │                               result
    │
    ▼
fmt.Println(string(output))
```

The JSON walker lives in `json_translator.go` (separated from `batch.go` during a refactor). It uses a fixed pool of 3 workers (`batchWorkers`). Maps are processed by sending key-value pairs over a buffered channel and writing results under a mutex. Slices are processed by sending indices over a channel — workers write directly to the slice by index, no mutex needed.

Each string translation goes through a semaphore (`sem chan struct{}` with cap 3) to cap concurrency at 3 in-flight requests to the backend. If any worker returns an error, it writes to `errCh` and cancels the shared context; all other workers see `ctx.Done()` and exit. Non-string values (numbers, booleans, null) pass through untouched with no function call.

## Language System

Languages are defined in a single global map:

```go
var languages = map[string]string{
    "auto": "Auto",
    "en":   "English",
    "it":   "Italian",
    // ... 25 languages total
}
```

At `init()` time, a sorted slice of codes is precomputed:

```go
func init() {
    codes := make([]string, 0, len(languages))
    for code := range languages { codes = append(codes, code) }
    sort.Strings(codes)
    langCodes = codes
}
```

`staticLanguages.List()` iterates `langCodes` and builds `[]Language` structs. Both the prompt builder (`defaultPrompt.Translate`) and the TUI's language selector read from the same map — no duplication. The benchmark tool in `cmd/bench` derives its target list from `NewStaticLanguages().List()` instead of maintaining a second copy.

## Configuration Loading

Config resolution is a cascade with two classes of paths:

```
--config <path>  ──► explicit  ──► must exist, error if missing
LOQI_CONFIG      ──► explicit  ──► must exist, error if missing
~/.config/loqi/config.yaml ──► optional ──► silently skip if missing
```

The `resolvePaths` function returns `(paths []string, explicit bool)`. If the caller specified a path (via flag or env var), `explicit` is `true` and `Load` errors on `ENOENT`. If using the default home-directory path, `explicit` is `false` and missing files are skipped.

The loaded YAML is unmarshalled into a pre-populated `Default()` struct, so partial configs work naturally:

```yaml
backend:
  base_url: http://192.168.1.100:11434
```

This changes only the URL; everything else keeps its default.

Options from `backend.options` are read as `map[string]any` and applied to the backend struct after construction. The helpers `intOption`, `floatOption`, and `durationOption` wrap the low-level `readFloatOption` to provide defaults.

## Ollama Lifecycle Management

`SetupOllama` in `translate/setup/server.go` coordinates three checks:

```
exec.LookPath("ollama")       ──► error if not installed
    │
ollama.Reachable(baseURL)     ──► GET /api/tags with 2s timeout
    │
    ├── reachable ──► skip start
    │
    └── not reachable ──► exec.Command("ollama", "serve")
                          WaitForReady(30, baseURL) — poll every 1s
                          timeout after 30s → kill process, error
    │
ollama.ModelExists(model, baseURL) ──► GET /api/tags, parse JSON, match name
    │
    ├── exists ──► skip pull
    │
    └── missing ──► PullModel(model, baseURL)
                     POST /api/pull with stream=true, 30min HTTP timeout
                     Line-by-line JSON scan → progress bar
                     error → kill Ollama if we started it
```

The `Reachable` check uses a shared package-level `httpClient` with 2-second timeout. `PullModel` uses a separate `pullClient` with 30-minute timeout because model downloads can be large. Progress rendering is in `progress.go` (separated from lifecycle logic during a refactor).

On cleanup, `UnloadModel` sends `POST /api/generate` with `keep_alive=0` to force Ollama to release the model — this prevents orphan `llama-server` processes from staying resident in memory.

## llama.cpp Lifecycle Management

`SetupLlamaCpp` in `translate/setup/server.go`:

```
llamacpp.ServerRunning(baseURL)   ──► GET /v1/models
    │
    ├── running ──► WaitForModelReady(60s) — poll /v1/models until 200
    │                 return (no process to kill on cleanup)
    │
    └── not running ──► exec.LookPath("llama-server")?
    │       │
    │       ├── not found ──► error
    │       │
    │       └── found + model_path set ──► exec.Command("llama-server",
    │                                        "--model", path,
    │                                        "--host", host,
    │                                        "--port", port,
    │                                        server_args...)
    │
    │                                      WaitForModelReady(60s)
    │                                      return (kill on cleanup)
    │
    └── not running + no model_path ──► error with instructions
```

Unlike Ollama, llama.cpp does not auto-pull models — it requires a local GGUF file. Extra flags (`--ctx-size`, `--ngl`, `--threads`, etc.) can be passed via the `server_args` config field.

## Argos Lifecycle Management

`SetupArgos` in `translate/setup/server.go`:

```
argos.Reachable(baseURL)      ──► TCP dial :5000 with 2s timeout
    │
    ├── reachable ──► skip start
    │
    └── not reachable ──► ensureVenv()
                              │
                              ├── venv created? ──► skip
                              │
                              └── no venv ──► python -m venv ~/.cache/loqi/argos-venv
                                              pip install argostranslate
                                                  │
                                                  ▼
                                          start embedded argos_server.py <port>
                                          wait up to 60s (poll every 500ms)
                                          timeout → kill process, error
```

The `ensureVenv` function creates a Python virtual environment in `~/.cache/loqi/argos-venv` (or `$TMPDIR/loqi-argos-venv` if home is unavailable). It looks for `python3` first, then falls back to `python` on Unix.

The embedded `argos_server.py` (bundled via `//go:embed`) is a lightweight HTTP server that wraps the `argostranslate` Python package. It accepts POST requests to `/translate` with `{q, source, target}` JSON and returns `{translatedText, error}`.

Argos does not auto-download language packages — `argostranslate` handles this internally on first use of a language pair. Subsequent translations reuse cached models. On cleanup, only the subprocess is killed if Loqi started it; there is no `UnloadBackend` call (no equivalent to Ollama's `keep_alive=0` endpoint).

**Known limitations:** does not support `--from auto` and requires Python 3 on the system. Also, first-run latency includes venv creation and pip install.

## Version Injection

A single variable `commands.Version` is injected at build time via `-ldflags`.
Both Makefile and goreleaser target the same symbol:

```makefile
# Makefile
LDFLAGS = -ldflags="-X github.com/danterolle/loqi/cmd/loqi/commands.Version=$(VERSION)"

# goreleaser
# -X github.com/danterolle/loqi/cmd/loqi/commands.Version={{ .Version }}
```

There is no runtime `git describe` call — it would fail in distributed binaries and was redundant given the Makefile and goreleaser both inject the tag at build time. On tag push (`v*.*.*`), the CI workflow runs goreleaser to produce platform binaries, then checks out `main`, runs `sed` to update the version badge in `docs/index.html`, and commits the change.

## Test Strategy

`translate.MockBackend` implements `translate.Backend` with a replaceable `TranslateFunc` field, defaulting to `"[source->target] text"`. Batch tests use it to verify JSON tree walking, structure preservation, non-string passthrough, and error propagation without real HTTP calls. Interface compliance is enforced at compile time with `var _ Backend = (*MockBackend)(nil)`.

Config tests validate defaults, file loading, partial overrides, and YAML parse errors.

The `tui` package has View-based tests that go through Bubble Tea's `Update()` message loop rather than calling internal methods. They verify that translation results render, stale data does not overwrite, errors show the right status while preserving output, and shortcuts like Ctrl+L and Tab work correctly.

The `commands` package has **no** test coverage.

## Known Limitations

- The batch worker pool is hardcoded to 3 goroutines with no configuration knob.
- There is no caching layer: every translation request, even for identical text, hits the backend API.
- `isThematicBreak` in `translate/markdown.go` matches any line of only `*`, `-`, `_`, or spaces (≥3 chars). This follows CommonMark but means any sequence of dashes and spaces like `- - -` is treated as a break, not a list. That is correct per spec, but could surprise users writing loose list markup.
- `splitPrefix` and its helpers (`splitWhitespace`, `splitAtxHeading`, `splitBlockquote`, `splitUnorderedList`, `splitOrderedList`) in `translate/markdown.go` index by byte, not rune. This is safe because all markdown prefixes (`#`, `>`, `-`, `*`, `+`, digits) are ASCII, but the mix of byte and `[]rune` in the same file is a maintenance trap if non-ASCII prefixes were ever added.
