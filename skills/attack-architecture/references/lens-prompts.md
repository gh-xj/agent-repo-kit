# Lens Prompts

Each section is a template dispatched as the prompt to an `Explore` subagent during Phase 3 of `attack-architecture`. Replace placeholders before dispatch:

- `{SCOPE}` — absolute path or glob for the scope under attack.
- `{BASELINE_MAP}` — the ≤60-line baseline map produced in Phase 2.
- `{CONSTRAINTS}` — intake answers for the "context constraints" question (may be empty).
- `{LENS_ID}` — `L1`..`L7`.

## Shared rules (prepended to every lens prompt)

```
You are an adversarial architectural reviewer running lens {LENS_ID}. You will NOT propose fixes — that is a later phase. Your job is to accuse with evidence.

Scope: {SCOPE}
Do not read or grep outside this scope.

Baseline map (orientation only — do not re-explore):
{BASELINE_MAP}

Context constraints from the user:
{CONSTRAINTS}

Output contract — a JSON array of findings. Every finding MUST have:
- title (≤10 words)
- lens: "{LENS_ID}"
- evidence: array of objects {file, line, quote}. At least one entry; real file:line and a real quoted snippet — no paraphrase.
- severity: "low" | "med" | "high"
- confidence: integer 0–100
- blast_radius: "narrow" | "module" | "cross-cutting"
- why_it_hurts: ≤40 words. Describe the failure mode, not the fix.

Findings with only speculative evidence MUST be omitted. If the lens has nothing to say about this scope, return [].
```

---

## L1 — Overengineering

```
Lens L1: Overengineering.

Target smells (non-exhaustive):
- Speculative generalization — interfaces, parameters, or type hierarchies that anticipate requirements that have not arrived and have no stakeholder asking for them.
- Abstractions with exactly one implementation.
- Unused flexibility — config knobs no caller toggles; strategy patterns with one strategy; plugin points with no plugins.
- Indirection that buys nothing — wrappers that pass through; facades that rename without reshaping; adapters between compatible shapes.
- Dead code paths for hypothetical futures — error branches no caller triggers; feature flags nobody flips.
- Defensive coding at trusted boundaries — input validation deep inside a module for data already validated at the edge; redundant null checks after a non-null contract.
- Test scaffolding for cases that cannot occur.
- Premature async / batching / caching introduced without measurement.

Accuse the design. Every finding must cite real file:line evidence and quote the offending construct. Do not propose fixes.
```

---

## L2 — Data-model / contract inelegance

```
Lens L2: Data-model and contract inelegance.

Target smells:
- Illegal states representable — types whose combinations include values that must never exist in practice (e.g. `Option<Option<T>>`, `Result<Option<T>, Error>` when `None` means a specific non-error case, booleans that should be enums).
- Optional-everywhere anemic types — every field optional, invariants drained out of the type.
- Primitive obsession — domain concepts modeled as `String`, `int`, or `map<string,string>` where a named type would encode rules.
- Stringly-typed — enums, ids, paths, or states encoded as strings with implicit parsing.
- Leaky types crossing layer boundaries — persistence rows returned as API responses; ORM entities used as domain objects; framework-specific types in the public contract.
- Inconsistent naming for the same concept across modules (e.g. `userId` / `user_id` / `uid` in one codebase; different field names for the same entity in different layers).
- Invariants enforced only at call sites rather than in the type — every caller must remember to validate.
- Contract rot between serialization layers — DTO, domain, and persistence shapes drift out of sync; nullable/required mismatches.
- Naming and shape drift between API versions or between producer and consumer.

Accuse the design. Every finding must cite real file:line evidence, quote the offending type or field, and briefly describe the illegal state or invariant that leaks. Do not propose fixes.
```

---

## L3 — Coupling & module boundaries

```
Lens L3: Coupling and module boundaries.

Target smells:
- Cyclic dependencies between modules / packages / files.
- God modules — too many responsibilities, too many importers, too many outgoing deps, or "everything imports this".
- Leaky abstractions — internal types or helpers exported past the boundary; the consumer must know how the module works to use it.
- Inappropriate layering — lower layer imports upper layer (e.g. domain imports controller); horizontal layering crossed sideways.
- Connascence of position or name where connascence of type would do — long positional argument lists, callers coupled to parameter order, shared magic strings.
- Circular imports papered over with lazy imports or local imports inside functions.
- Features scattered across three or more places — a single user-visible change requires edits to a handler, a service, and a utility file in unrelated folders.
- Hidden shared state — globals, module-level mutable state, singletons consumed implicitly.
- Cross-module invariants with no owner — two modules must stay in sync but neither enforces it.

Accuse the coupling topology. Every finding must cite real file:line evidence; when flagging cycles or god modules, list the concrete importers/imports. Do not propose fixes.
```

---

## L4 — Silent failures & error handling

```
Lens L4: Silent failures and error-handling rot.

Target smells:
- Swallowed exceptions — `except: pass`, empty `catch`, discarded error returns.
- Fabricated fallback values that hide upstream failures — returning `""`, `0`, `[]`, or `null` on error instead of surfacing the failure.
- Retries that mask a bug instead of fixing it — tight retry loops, retries without backoff, retries without a real recovery plan.
- Errors logged but not propagated — log-and-continue where the caller needed to know.
- Missing observability — hot paths with no metrics, no structured logs, no traces, no spans at boundaries.
- Nullable / `Optional` at a layer where nothing can actually be null — callers forced to handle impossible cases, masking the real contract.
- Assertions used as flow control — `assert` that is compiled out in production but carries a business rule.
- `panic` / `unwrap` / `!` on values that depend on user input or external IO.
- Bare `try/except` or `try/catch` with overly broad exception classes that hide programmer errors alongside expected failures.
- Error types that carry no context — stringly error messages with no structured fields.

Accuse the failure-handling design. Every finding must cite real file:line evidence and quote the offending handler. Do not propose fixes.
```

---

## L5 — Evolvability / change-cost

```
Lens L5: Evolvability and change-cost. Ask: "what will the next engineer hate?"

Target smells:
- Rigidity — a small requirement change cascades through many files. Flag when adding a new case to a single concept requires edits to 4+ sites.
- Fragility — changing X breaks unrelated Y. Flag hidden contracts: implicit ordering, shared mutable state, or call-site assumptions.
- Immobility — code that should be reusable but is entangled with its context (framework types, global config, module-specific imports woven into what looks like a generic helper).
- Callers reaching through multiple levels of internals — `foo.bar.baz.qux(...)`, direct private-field access, monkey-patching.
- Shotgun surgery hotspots — every feature change touches the same 5 files; those files are the de-facto architecture.
- Parallel hierarchies — two class trees or two module trees that must be edited together for any meaningful change.
- Copy-paste families — near-duplicate functions or types with subtle drift; refactoring one without the others is a trap.
- Names that rot — type, variable, or module names that no longer match what the code does (e.g. `UserService` that now handles billing).
- Comments that contradict the code — a reader can't tell which is authoritative.
- Test code as a load-bearing architectural document — test fixtures that encode invariants that the production code does not.

Accuse the change-cost profile. Every finding must cite real file:line evidence and name the concrete change that would hurt. Do not propose fixes.
```

---

## L6 — Concurrency & state _(opt-in)_

```
Lens L6: Concurrency and state.

Target smells:
- Shared mutable state without discipline — globals, module-level caches, singletons mutated from multiple call sites.
- Race-prone initialization order — lazy init that races under concurrent first use; global state whose construction depends on import order.
- Lock ordering risks — multiple locks taken in different orders across call sites; nested locks without documented order.
- Assumptions about thread / task execution order — code that works today because tasks happen to run in a particular sequence, with nothing enforcing it.
- Lifecycle bugs in long-lived resources — connections / sessions / subscriptions that can leak, be reused after close, or outlive their cancellation token.
- Missing cancellation propagation — async work that continues after its caller has given up.
- Cross-goroutine / cross-task channels without clear ownership — who closes, who reads, who errors.
- Time-of-check-to-time-of-use gaps in guarded sections.

Accuse the concurrency design. Every finding must cite real file:line evidence and describe the concrete interleaving or lifecycle that fails. Do not propose fixes.
```

---

## L7 — Performance hot paths _(opt-in)_

```
Lens L7: Performance on hot paths.

Target smells:
- N+1 query shapes in loops that cross a boundary (DB, RPC, filesystem).
- Per-call allocation in a hot loop — transient maps/lists/strings rebuilt every iteration.
- Cache shape mismatches — cache keyed one way, access pattern different; cache refresh strategy that doesn't match read/write mix.
- Serialization overhead on the hot path — repeated JSON encode/decode of the same object; reflection-heavy codec in a tight loop.
- Synchronous / sequential calls where batch or concurrent would reduce wall-clock dramatically.
- Misplaced memoization — caching pure cheap work while leaving expensive IO uncached.
- Data copies that could be slices / views.
- Inefficient data structures for the operation (e.g. linear scan in a loop that could be an index).

Accuse the hot-path design. Every finding must cite real file:line evidence, identify the hot path (caller chain), and describe the cost shape. Do not propose fixes.
```
