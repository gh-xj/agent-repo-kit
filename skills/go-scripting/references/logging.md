# Logging

## zerolog vs slog+tint

| Axis                                | `rs/zerolog`                                                   | `log/slog` + `lmittmann/tint`                           |
| ----------------------------------- | -------------------------------------------------------------- | ------------------------------------------------------- |
| Status                              | Third-party, mature (since 2017)                               | Stdlib since Go 1.21 (2023) + single-file tint handler  |
| API shape                           | Fluent chain: `log.Info().Str("k","v").Msg("hi")`              | Key-value varargs: `slog.Info("hi", "k", v)`            |
| Perf                                | Zero-alloc, famously fast                                      | Fast; not zero-alloc                                    |
| Levels                              | Trace / Debug / Info / Warn / Error / Fatal / Panic / Disabled | Debug / Info / Warn / Error                             |
| Pretty console                      | `zerolog.ConsoleWriter`                                        | `tint.NewHandler`                                       |
| Dep footprint                       | 1 direct (zerolog itself pulls nothing)                        | 1 direct (tint), otherwise stdlib                       |
| Context plumbing                    | Works, but not ecosystem-wide                                  | `slog.With`, `slog.FromContext`-style idioms everywhere |
| Handler swap (JSON / text / custom) | Writer-based, less composable                                  | First-class Handler interface; swap freely              |
| Community direction                 | Stable, mature, slow-moving                                    | Where new Go code is consolidating                      |

## Pick slog+tint. Reasons, in order

1. **Stdlib beats third-party when the delta is small.** Dependencies outlive their utility; stdlib doesn't rot.
2. **Ecosystem momentum.** Since Go 1.21 (Aug 2023), new libraries expect `slog`. You get free `context.Context` propagation, free handler swap, free JSON mode, free integration with any library that speaks slog.
3. **Simplicity.** `slog.Info("linked skill", "name", skill)` is a single line; zerolog's chain is fine but adds visual ceremony.
4. **Zero-alloc is not your bottleneck.** In a CLI the I/O and the thing being done dominate. Save micro-optimization for hot-path services that have measured.
5. **Fewer decisions.** slog's level set is a boring {Debug,Info,Warn,Error}. zerolog's eight levels tempt over-categorization.

## Pick zerolog only if

- You're working in a codebase that already uses it (don't mix loggers).
- You're in a service hot path where allocation truly matters AND you've measured.
- You need a level between Debug and Info (Trace). Rare in CLIs.

For new Go scripts and tools: **slog + tint**.

## Setup boilerplate

Drop this into `internal/log/log.go`:

```go
package log

import (
	"io"
	"log/slog"
	"os"

	"github.com/lmittmann/tint"
	"golang.org/x/term"
)

type Options struct {
	Verbose bool // --verbose → Debug level
	NoColor bool // --no-color forces plain output
	Writer  io.Writer
}

func Setup(opts Options) {
	w := opts.Writer
	if w == nil {
		w = os.Stderr
	}
	noColor := opts.NoColor || os.Getenv("NO_COLOR") != "" || !isTerminal(w)
	level := slog.LevelInfo
	if opts.Verbose {
		level = slog.LevelDebug
	}
	handler := tint.NewHandler(w, &tint.Options{
		Level:   level,
		NoColor: noColor,
		// Strip timestamp — CLI output shouldn't be noisy.
		ReplaceAttr: func(_ []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey {
				return slog.Attr{}
			}
			return a
		},
	})
	slog.SetDefault(slog.New(handler))
}

func isTerminal(w io.Writer) bool {
	f, ok := w.(*os.File)
	if !ok {
		return false
	}
	return term.IsTerminal(int(f.Fd()))
}
```

Call `log.Setup(...)` once near program start, before any subcommand runs.

## Idioms

- **One line per event.** `slog.Info("did thing", "key", value, "key2", value2)`.
- **Data in attrs, not in the message.** Prefer `"linked skill", "name", skill` over string concatenation.
- **Respect `NO_COLOR`.** tint's `NoColor` option covers that plus non-TTY detection.
- **Strip timestamps for CLIs.** Your terminal already has a prompt timestamp. Drop them via `ReplaceAttr`. Re-enable for daemons or when piping to a log aggregator.
- **Don't log at multiple levels for the same event.** Pick one.
- **Context plumbing:** build request-scoped loggers with `slog.With(...)` and pass them around, not the handler.

## Anti-patterns

- **`log.Fatal` / `slog.Error` + `os.Exit` inside library code.** Return errors; exit from `main`.
- **Multiple loggers across one binary.** Set one default in `Setup` and call `slog.Info` / `slog.Warn` / `slog.Error` everywhere.
- **Ad-hoc ANSI escapes.** Use tint or a structured handler; don't hand-roll `\033[31m`.
- **Structured attrs mixed with interpolation.** Don't write `slog.Info(fmt.Sprintf("hi %s", x))` — write `slog.Info("hi", "target", x)`.
