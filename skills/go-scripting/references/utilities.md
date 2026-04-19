# Utility Libraries: `samber/lo`

## The honest take

[`samber/lo`](https://github.com/samber/lo) is a Go port of JavaScript's lodash. It offers `Map`, `Filter`, `Reduce`, `GroupBy`, `Chunk`, `Uniq`, `Partition`, and ~100 more generic helpers.

It is a legitimately useful library. It is also legitimately contested in Go, because Go's idiom has always been "a little copying is better than a little dependency" and "explicit loops over higher-order abstractions for readability."

This guide prescribes a **middle path**: use `samber/lo` when it earns its keep, not as a blanket replacement for `for`.

## When it earns its keep

Reach for `samber/lo` when **any** of:

- **3+ functional ops compose** on a collection: `lo.Uniq(lo.Map(items, f))` beats the equivalent nested loop by readability.
- Stdlib doesn't have it cleanly: `lo.GroupBy`, `lo.Chunk`, `lo.Partition`, `lo.Keyify`, `lo.SliceToMap`.
- You're writing a **data transformation pipeline** where the intent reads more clearly as a chain than as loops.
- You need **typed generic helpers** without rolling them yourself: `lo.Ternary`, `lo.CoalesceOrEmpty`, `lo.ToPtr` (but see "nil-handling" below).

## When it doesn't

### Single Filter or Map

The loop is clearer and zero-dep:

```go
// Write this:
var evens []int
for _, n := range nums {
    if n%2 == 0 {
        evens = append(evens, n)
    }
}

// Not this:
evens := lo.Filter(nums, func(n int, _ int) bool { return n%2 == 0 })
```

The loop is roughly the same length once you count the import and the closure ceremony, and it says exactly what's happening with no indirection.

### Hot paths

`lo` closures allocate per call and are not inlined across the generic boundary. If you're inside a tight loop processing millions of items, write the range loop.

### Code readers unfamiliar with `lo`

If your team doesn't use `lo` elsewhere, don't introduce it for one function. New deps are a team-level decision.

### Dangerously "helpful" helpers

`lo` has some helpers whose semantics are subtly un-Go-ish:

- `lo.Must(err)` — panics on error. Use sparingly; never in library code.
- `lo.Ternary(cond, a, b)` — **both branches evaluate**. A plain `if`/`else` short-circuits.
- `lo.ToPtr(x)` — fine, but a 1-line helper in your own package is often clearer and avoids the dep spread.

## Rule of thumb

- **First site** of a helper: **write the loop**.
- **Second site** (same helper, different place): **copy the loop, add a `TODO: dedupe`** if you want.
- **Third site:** either extract a package-local helper, or reach for `lo` if the operation is a pure FP primitive.

Don't import a dep to save three lines in one place.

## Stdlib alternatives

Check stdlib first. Since Go 1.21 / 1.22:

- `slices` — `slices.Contains`, `slices.Index`, `slices.Delete`, `slices.Sort`, `slices.SortFunc`, `slices.Equal`, `slices.Reverse`, `slices.Max`, `slices.Min`.
- `maps` — `maps.Keys`, `maps.Values`, `maps.Clone`, `maps.Copy`, `maps.Equal`.
- `cmp` — `cmp.Compare`, `cmp.Less`, `cmp.Or`.
- `iter` (Go 1.23+) — range-over-func for user-defined iterators.

Between `slices`, `maps`, `cmp`, and `iter`, roughly **40% of what people reach for `lo` for is now stdlib**. Prefer stdlib when available.

## Anti-patterns

- **`lo.Map` + `lo.Filter` + `lo.Uniq` on a slice of 5.** Write the loop.
- **`lo.Must(json.Unmarshal(...))` in production code.** Panics are a bad user experience.
- **Mixing `lo` and stdlib `slices` in the same file.** Pick one surface.
- **`lo.Ternary` for side-effecting branches.** Both sides run — use `if`.
