# Error Messages as Remediation

Error messages are the single most leveraged agent-context surface you
control. When a lint, CLI, or test fails, the message is what the agent reads
next. Treat it as an **instruction**, not a diagnosis.

## Invariant

Every error message a tool emits MUST state:

1. **What went wrong** (the current behaviour).
2. **What to do next** (a specific, actionable fix).

A message that states only #1 is half-finished.

## Shape (invariant, not template)

The fix can live inside the same line (`→ fix: ...`), on a following line, or
in a structured field. The shape is the author's call. The constraint is that
an agent reading the output knows the next verb to run or the next edit to
make without further inference.

## Why

From [[S-001]]: "Because the lints are custom, we write the error messages
to inject remediation instructions into agent context." In agent-first
systems, error messages are dominant bandwidth. A terse error forces the
agent (or human) to consult a separate doc, skim source, or guess. A
remediation-rich error closes the loop in-place.

## Examples from this repo (evidence)

Before (terse):

```
invalid priority: P5
allowed: P0, P1, P2, P3
```

After (remediation-rich):

```
invalid priority: P5  → fix: retry with --priority <value>
allowed: P0, P1, P2, P3
```

Before:

```
wikilink [[S-404]] has no target page
```

After:

```
wikilink [[S-404]] has no target page  → fix: create 'pages/S-404-<slug>.md', remove the link, or fix the typo
```

Before:

```
unknown view
```

After:

```
unknown view "ready-now"  → fix: run 'work view ready' or choose one of the built-in views
```

Implementations: the `work` CLI, `.wiki/scripts/lint.sh`, and the template
copies under `references/templates/`.

## Anti-patterns

- **Stacktrace without remediation.** Raw stack traces are diagnostic, not
  instructive. If the tool knows the fix, say it. If it doesn't, document
  the most likely fix based on error class.
- **"See the docs."** If the docs are the fix, inline the relevant sentence
  from them. Indirection costs a round-trip.
- **Listing all possible causes.** One good guess at the likely fix beats
  an exhaustive decision tree. If ambiguity is high, list the top two.
- **Generic validation errors.** `invalid input` with no context forces
  re-reading the call site. Name the specific field and the expected shape.

## When NOT to inline remediation

- **Programmatic consumers** (e.g. an API where the caller parses structured
  errors). Use error codes + a documented remediation map, not free text.
- **Security-sensitive surfaces** where the fix would leak internal state
  (e.g., a login error that says "user X doesn't exist" is a remediation
  that's also a user-enumeration vulnerability).
- **Truly novel errors.** If the tool has never seen this case before, honest
  "unknown error, please report" beats a wrong hint.

## Review rule

When reviewing a PR that adds or changes error messages, ask: _would an
agent with zero context for this repo know what to do after reading this
message alone?_ If no, push back.
