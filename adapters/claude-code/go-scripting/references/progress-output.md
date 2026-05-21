# Progress and Output

Two common output needs that CLIs grow into: **tables** for structured
multi-row data and **spinners/progress** for long-running operations.
Both have sharp edges around non-TTY output — get that right first.

## Non-TTY first

A CLI runs in three modes:

1. **Interactive TTY** — colors, spinners, progress bars, tables.
2. **Piped to another command or file** — plain text, no ANSI escapes,
   no carriage-return redraws.
3. **CI / log aggregator** — plain text, structured ideally, no spinners.

Every visual library must respect `NO_COLOR`, `term.IsTerminal`, and
`--no-color`. If you emit spinners and tables into a log file, you
produce garbage.

```go
import "golang.org/x/term"

interactive := term.IsTerminal(int(os.Stdout.Fd())) &&
	os.Getenv("NO_COLOR") == ""
```

Gate every spinner / progress bar / color-rich table behind `interactive`.

## Tables

Use a table when the output is **row-oriented structured data** that a
human will scan visually. Do **not** use a table for one-of-a-kind
summaries (use log lines) or for machine consumption (use JSON, gated by
`--json`).

### jedib0t/go-pretty

[`jedib0t/go-pretty/v6/table`](https://github.com/jedib0t/go-pretty) is the
modern go-to: rich styles, colored cells, row grouping, pagination,
Unicode-aware width. Two dependencies, one of which is itself.

```go
import "github.com/jedib0t/go-pretty/v6/table"

t := table.NewWriter()
t.SetOutputMirror(os.Stdout)
t.AppendHeader(table.Row{"#", "name", "status"})
t.AppendRows([]table.Row{
	{1, "skill-builder", "ok"},
	{2, "convention-engineering", "ok"},
})
t.SetStyle(table.StyleRounded)
t.Render()
```

For a simpler dependency, `olekukonko/tablewriter` works but is
older-feeling and less actively maintained.

### When NOT to use a table

- Output piped to `grep` / `awk` — emit plain columns or JSON.
- One row with 2–3 fields — just log the fields.
- `--json` flag is set — emit JSON, skip the table.

## Spinners

Use a spinner when:

- An operation takes 3+ seconds with no intermediate output.
- The user otherwise has no signal that something is happening.

Don't use a spinner when:

- The operation has useful per-step output (log it instead).
- Output is being piped (detect and suppress).
- You're in CI or a log file (detect and suppress).

### charmbracelet/huh/spinner

[`charmbracelet/huh/spinner`](https://github.com/charmbracelet/huh) — pairs
well with huh forms; consistent Charm aesthetic; light API:

```go
import "github.com/charmbracelet/huh/spinner"

_ = spinner.New().
	Title("Downloading…").
	Action(func() {
		download(ctx)
	}).
	Run()
```

Blocking call; swaps back to normal output when `Action` returns.

### briandowns/spinner

[`briandowns/spinner`](https://github.com/briandowns/spinner) — older but
minimal, no Charm ecosystem commitment:

```go
import "github.com/briandowns/spinner"

s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
s.Suffix = " Downloading…"
s.Start()
defer s.Stop()
download(ctx)
```

Start/Stop-style API. Good when you need to update the message mid-flight.

### Rule of thumb

- Using Charm/huh already → `huh/spinner`.
- Plain CLI, want the smallest footprint → `briandowns/spinner`.
- Must update spinner text repeatedly → `briandowns/spinner` (has `Suffix`
  you can assign while running).

## Progress bars (when you have work units)

If you know the total (bytes to download, files to process), use a
progress bar, not a spinner:

- [`schollz/progressbar/v3`](https://github.com/schollz/progressbar) —
  minimal, single-bar, good default choice.
- Charm ecosystem has `bubbles/progress` (part of bubbletea); use that if
  you're already building a TUI.

## JSON-or-table decision

Support `--json` on any command that emits structured data. The code path:

```go
if cli.JSON {
	return json.NewEncoder(os.Stdout).Encode(results)
}
renderTable(results)
```

Never emit both. JSON is for pipelines; table is for humans. A flag
decides; never sniff the terminal to pick.

## Anti-patterns

- **Spinners in CI logs.** Each frame is a new line; your CI output
  becomes unreadable. Detect `!term.IsTerminal || NO_COLOR` and suppress.
- **Tables emitted when `--json` was passed.** Respect the flag.
- **Colored output without `NO_COLOR` support.** Contract violation.
- **Carriage-return redraws in piped output.** Detect TTY; skip.
- **Progress bar for a <1s operation.** Distraction, not help.
- **Two bars/spinners at once.** They'll interleave and glitch. Serialize
  them, or use a multi-progress library.
