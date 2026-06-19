## What It Is

Voca is a terminal-based translator that runs locally through Ollama LLMs. It works in three modes: an interactive TUI built with bubbletea, a `voca translate` command for one-shot translations, and `voca batch` for bulk-translating JSON or plain text files. 

The entire codebase is Go with only five external dependencies (bubbletea, bubbles, lipgloss, yaml.v3, atotto/clipboard), and the sole backend is Ollama over HTTP.

## Package Structure

Seven packages (plus `cmd/bench`) with a linear dependency graph — no cycles:

```
cmd/voca/main.go
  └─ cmd/voca/commands/     ── orchestrates everything
       ├─ app.go            ── dispatch, usage, flag parsing
       ├─ translate.go      ── RunTranslate, RunCLI
       ├─ batch.go          ── RunBatch
       ├─ tui.go            ── RunTUI
       ├─ setup.go          ── backend initialization
       ├─ ollama.go         ── Ollama lifecycle (start, wait, pull)
       ├─ io.go             ── input reading
       └─ banner.go         ── ANSI logo
  ├─ translate/             ── domain logic
  │   ├─ interfaces.go      ── Backend, PromptBuilder, LanguageProvider
  │   ├─ core.go            ── thin Backend + LanguageProvider wrapper
  │   ├─ languages.go       ── language map + sorted codes (init-time)
  │   ├─ default_prompt.go  ── system + user prompt templates
  │   ├─ batch.go           ── recursive JSON walker + worker pool
  │   ├─ mock_backend.go    ── test double
  │   └─ ollama/
  │       ├─ backend.go     ── HTTP /api/chat client
  │       └─ lifecycle.go   ── health checks, model pull
  ├─ tui/                   ── Bubble Tea app
  │   ├─ model.go / update.go / view.go
  │   ├─ commands.go        ── doTranslate, copyClipboard
  │   ├─ styles.go / ui.go
  ├─ config/                ── YAML config loader
  └─ cmd/bench/             ── multi-language benchmark harness
```

Domain code lives in `translate` with its interfaces; `commands` handles setup and dispatch; `tui` owns the UI; `config` loads and merges YAML.

## TUI Mode

When the user launches `voca` with no arguments, `Run()` falls through to `RunTUI`, which calls `setupRun` to initialize the backend and then passes `core.Backend` and `core.Languages` directly to `RunBubbleTea` — the TUI has no dependency on `Core` itself.

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
    │       │           doTranslate(text) ──► HTTP /api/chat ──► parse response
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
                         setupRun(cfg, model)
                             │
                             ├── printBanner()
                             ├── SetupOllama(model, baseURL)
                             │       ├── Reachable? ──► no ──► start ollama serve
                             │       │                      ──► WaitForReady(30s)
                             │       ├── ModelExists? ──► no ──► PullModel
                             │       └── return cmd handle
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

The signal context ensures that if the user presses CTRL+C while translating, the deferred `cleanup()` runs — which kills the Ollama process only if Voca started it. This distinction matters: if Ollama was already running when Voca launched, cleanup is a no-op.

`setupRun` is the central factory. It prints the banner, ensures Ollama is running and has the model, creates the HTTP backend with config-driven overrides (`temperature`, `top_p`, `num_predict`, `timeout`), and returns a `*translate.Core` together with a cleanup closure. 

Both `RunTranslate` and `RunBatch` call it, as does `RunTUI`. The `translate.Core` it returns is a thin struct that aggregates a `Backend` and a `LanguageProvider` — its `Translate` method simply delegates to `Backend.Translate`, and the TUI ignores it entirely, using `core.Backend` and `core.Languages` directly.

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
    │       │                               │
    │       │                               ▼
    │       │                      json.MarshalIndent(data)
    │       │                               │
    │       │                               ▼
    │       │                             result
    │       │
    │       └── no ──► core.Translate(ctx, text, from, to)
    │                                  │
    │                                  ▼
    │                               result
    │
    ▼
fmt.Println(string(output))
```

The JSON walker uses a fixed pool of 3 workers (`batchWorkers`). Maps are processed by sending key-value pairs over a buffered channel and writing results under a mutex. Slices are processed by sending indices over a channel — workers write directly to the slice by index, no mutex needed.

```
processMapNode:
    collect entries ──► buffered channel ──► 3 goroutines ──► processNode each child
                                                       │
                                                       ▼
                                              mutex protect map write

processSliceNode:
    collect indices ──► buffered channel ──► 3 goroutines ──► processNode each child
                                                       │
                                                       ▼
                                              direct slice[i] write
```

Each string translation goes through a semaphore (`sem chan struct{}` with cap 3) to cap concurrency at 3 in-flight requests to Ollama. If any worker returns an error, it writes to `errCh` and cancels the shared context; all other workers see `ctx.Done()` and exit. Non-string values (numbers, booleans, null) pass through untouched with no function call.

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

Options from `backend.options` are read as `map[string]any` and applied to the `ollama.Backend` struct after construction. The `readFloatOption` helper handles both `float64` and `int` YAML types, since the YAML parser returns `int` for unquoted integers and `float64` for decimal numbers.

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

The `Reachable` check uses a shared package-level `httpClient` with 2-second timeout. `PullModel` uses a separate `pullClient` with 30-minute timeout because model downloads can be large. Progress rendering happens in `renderPullStatus` which paints an ANSI progress bar for download percentages, short status lines for layer pulling, and a checkmark on completion.

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

- The `wrap()` function in `tui/view.go` splits on spaces — it does not handle CJK text where word boundaries are not marked by whitespace so Chinese, Japanese and Korean output will not wrap correctly in the TUI output pane.
- The batch worker pool is hardcoded to 3 goroutines with no configuration knob.
- Only the Ollama backend exists; the `Backend` interface would accept others, but none are implemented.
- At the moment there is no caching layer: every translation request, even for identical text, hits the Ollama API.
