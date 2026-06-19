## What It Is

Voca is a terminal-based translator that runs locally through LLMs. It supports two backends: **Ollama** (default) and **llama.cpp**. It works in three modes: an interactive TUI built with bubbletea, a `voca translate` command for one-shot translations, and `voca batch` for bulk-translating JSON or plain text files.

The entire codebase is Go with only five external dependencies (bubbletea, bubbles, lipgloss, yaml, atotto/clipboard).

## Package Structure

Eight packages (plus `cmd/bench`) with a linear dependency graph — no cycles:

```
cmd/voca/main.go
  └─ cmd/voca/commands/     ── orchestrates everything
       ├─ app.go            ── dispatch, usage, flag parsing
       ├─ translate.go      ── RunTranslate, RunCLI
       ├─ batch.go          ── RunBatch
       ├─ tui.go            ── RunTUI
       ├─ setup.go          ── SetupRun (backend routing), option helpers
       ├─ ollama.go         ── Ollama lifecycle (start, wait, pull)
       ├─ llamacpp.go       ── llama.cpp lifecycle (start, wait)
       ├─ io.go             ── input reading (args, file, stdin)
       └─ banner.go         ── ANSI logo
  ├─ translate/             ── domain logic
  │   ├─ interfaces.go      ── Backend, PromptBuilder, LanguageProvider
  │   ├─ core.go            ── thin Backend + LanguageProvider wrapper
  │   ├─ languages.go       ── language map + sorted codes (init-time)
  │   ├─ default_prompt.go  ── system + user prompt templates
  │   ├─ batch.go           ── batch entry point (JSON dispatch)
  │   ├─ json_translator.go ── recursive JSON walker + worker pool
  │   ├─ mock_backend.go    
  │   ├─ ollama/
  │   │   ├─ backend.go     ── HTTP /api/chat client
  │   │   ├─ lifecycle.go   ── health checks, model pull/unload
  │   │   └─ progress.go    ── ANSI progress bar rendering
  │   └─ llamacpp/
  │       ├─ backend.go     ── OpenAI-compatible /v1/chat/completions client
  │       └─ lifecycle.go   ── server check, model-ready polling
  ├─ tui/                   ── Bubble Tea app
  │   ├─ model.go / update.go / view.go
  │   ├─ commands.go        ── doTranslate, copyClipboard
  │   ├─ styles.go / ui.go
  ├─ config/                ── YAML config loader
  └─ cmd/bench/             ── multi-language benchmark
```

Domain code lives in `translate` with its interfaces; `commands` handles setup and dispatch; `tui` owns the UI; `config` loads and merges YAML.

## Backend Selection

SetupRun dispatches based on `cfg.Backend.Type`:

- **`ollama`** — calls `SetupOllama` (starts `ollama serve` if not running, pulls model if missing, calls `UnloadModel` on cleanup)
- **`llamacpp`** — calls `SetupLlamaCpp` (starts `llama-server --model <path>` if `model_path` is set, or connects to an existing server; no auto-pull; kills subprocess on cleanup if Voca started it)

Both paths return a `*translate.Core` wrapping a backend that satisfies `translate.Backend`, plus a `func()` cleanup closure.

```go
type Backend interface {
    Translate(ctx context.Context, text, source, target string) (string, error)
}
```

## TUI Mode

When the user launches `voca` with no arguments, `Run()` falls through to `RunTUI`, which calls `SetupRun` to initialize the backend and then passes `core.Backend` and `core.Languages` directly to `RunBubbleTea` — the TUI has no dependency on `Core` itself.

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
    headerView    ──► "voca  From: Italian  ->  To: English"
    textarea.View ──► input area
    outputView    ──► wrapped translation
    statusView    ──► "Ready.  ctrl+y:copy  ctrl+l:clear  ..."
```

The first keystroke translates immediately (`leadingDone` gate). Every subsequent keystroke increments `translateSeq` and schedules a debounce tick. If a new keystroke arrives before the tick fires, the old tick is ignored because its sequence number no longer matches. When the result arrives, it is compared against the current textarea value: if the user changed the input while waiting, the result is thrown away. This prevents the classic race where a slow response overwrites a newer translation.

The `lastInput` field exists to solve a subtle bug: without it, the debounce handler compared `m.output` (the previous translation result) against `m.textarea.Value()` (the new input). Those are different domains — input text vs. translated text — so the comparison would miss real changes. Now it compares the current input against the last input that was actually sent for translation, which is the correct check.

## CLI Mode

`voca translate --from it --to en "Ciao mondo"` takes a simpler path:

```
parseTranslateFlags ──► ReadInput (text, file or stdin)
                             │
                             ▼
                         SetupRun(cfg, model)
                             │
                             ├── printBanner()
                             ├── switch cfg.Backend.Type:
                             │     ollama  ──► SetupOllama()
                             │                   ├── Reachable? ──► no ──► start ollama serve
                             │                   │                      ──► WaitForReady(30s)
                             │                   ├── ModelExists? ──► no ──► PullModel
                             │                   └── return cmd handle
                             │     llamacpp ──► SetupLlamaCpp()
                             │                    ├── ServerRunning? ──► yes ──► wait for model
                             │                    ├── no + model_path? ──► start llama-server
                             │                    └── return cmd handle
                             ├── build backend with config options
                             └── return *Core + cleanup()
                             │
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
                         fmt.Println(result)
```

The signal context ensures that if the user presses CTRL+C while translating, the deferred `cleanup()` runs — which kills the subprocess only if Voca started it. This distinction matters: if the backend was already running when Voca launched, cleanup is a no-op.

## Batch Mode

`voca batch --from en --to it < locales/en.json` handles JSON and plain text differently:

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
VOCA_CONFIG      ──► explicit  ──► must exist, error if missing
~/.config/voca/config.yaml ──► optional ──► silently skip if missing
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

`SetupOllama` in `commands/ollama.go` coordinates three checks:

```
exec.LookPath("ollama")           ──► error if not installed
    │
ollama.Reachable(baseURL)         ──► GET /api/tags with 2s timeout
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

`SetupLlamaCpp` in `commands/llamacpp.go`:

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
    │                                      WaitForModelReady(60s)
    │                                      return (kill on cleanup)
    │
    └── not running + no model_path ──► error with instructions
```

Unlike Ollama, llama.cpp does not auto-pull models — it requires a local GGUF file. Extra flags (`--ctx-size`, `--ngl`, `--threads`, etc.) can be passed via the `server_args` config field.

## Version Injection

Version comes exclusively from `-ldflags` at build time:

```makefile
# goreleaser sets this via ldflags:
# -X github.com/danterolle/voca/cmd/voca/commands.Version={{ .Version }}
```

There is no runtime `git describe` call — it would fail in distributed binaries and was redundant given the Makefile and goreleaser both inject the tag at build time. On tag push (`v*.*.*`), the CI workflow runs goreleaser to produce platform binaries, then checks out `main`, runs `sed` to update the version badge in `docs/index.html`, and commits the change.

## Test Strategy

`translate.MockBackend` implements `translate.Backend` with a replaceable `TranslateFunc` field, defaulting to `"[source->target] text"`. This lets the batch tests verify JSON tree walking, structure preservation, non-string passthrough, and error propagation without any HTTP calls.

Config tests verify defaults, file loading, partial overrides, and YAML parse errors. Interface compliance is checked at compile time with package-level `var _ Backend = (*MockBackend)(nil)` assertions.

There is no test coverage for the `tui` or `commands` packages.

## Known Limitations

- The `wrap()` function in `tui/view.go` splits on spaces — it does not handle CJK text where word boundaries are not marked by whitespace, so Chinese, Japanese and Korean output will not wrap correctly in the TUI output pane.
- The batch worker pool is hardcoded to 3 goroutines with no configuration knob.
- There is no caching layer: every translation request, even for identical text, hits the backend API.
