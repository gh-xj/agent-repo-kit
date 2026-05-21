# Collection Patterns

Modern Go gives you four well-ordered layers for working with slices, maps,
and sequences. Pick the lowest layer that cleanly expresses what you want.

## Priority order

1. **Stdlib `slices` / `maps` / `cmp` (Go 1.21+)** — first reach. Typed,
   zero-dep, ships with the toolchain.
2. **Stdlib `iter` + range-over-func (Go 1.23+)** — when you need lazy
   iteration, composition, or to expose a sequence without building a slice.
3. **Plain `for` loop.** Always fine. Sometimes clearest.
4. **`samber/lo`** — last resort, for specific helpers the first three
   layers don't cleanly provide.

## Layer 1: stdlib `slices` / `maps` / `cmp`

Since Go 1.21, stdlib covers the common generic operations:

```go
import "slices"
import "maps"
import "cmp"

// slices
slices.Contains(xs, target)
slices.Index(xs, target)
slices.Sort(xs)                       // in place
slices.SortFunc(xs, func(a, b T) int { return cmp.Compare(a.K, b.K) })
slices.Equal(xs, ys)
slices.Reverse(xs)
slices.Max(xs), slices.Min(xs)
slices.Delete(xs, i, j)

// maps
maps.Keys(m)       // returns iter.Seq[K] in 1.23+ (was []K in 1.21-1.22)
maps.Values(m)
maps.Clone(m)
maps.Copy(dst, src)
maps.Equal(m1, m2)

// cmp
cmp.Compare(a, b)  // -1 / 0 / 1
cmp.Or(a, b, c)    // first non-zero; useful for default fallbacks
```

Go 1.23 adds `slices.Chunk`, `slices.All`, `slices.Values`, `slices.Sorted`,
`slices.Collect` — most of which return or consume `iter.Seq` so they
integrate with Layer 2.

## Layer 2: `iter` + range-over-func (Go 1.23+)

`iter.Seq[V]` is `func(yield func(V) bool)` — a pull-style iterator over
zero or more values. Range-over-func lets you iterate with regular
`for … range`:

```go
import "iter"

func EvenNumbers(xs []int) iter.Seq[int] {
	return func(yield func(int) bool) {
		for _, x := range xs {
			if x%2 == 0 {
				if !yield(x) {
					return
				}
			}
		}
	}
}

for x := range EvenNumbers(nums) {
	fmt.Println(x)
}
```

`iter.Seq2[K, V]` is the key-value variant (used by `maps.All`,
`maps.Keys`, `maps.Values`).

### Where iter wins over lo

Anywhere you were reaching for `lo.Map(lo.Filter(...))` to lazily compose:

```go
// Was: lo.Uniq(lo.Map(lo.Filter(items, isReady), renderID))
// Now: compose iterators; no intermediate slices

func Ready(items []Item) iter.Seq[Item] { ... }
func Rendered(items iter.Seq[Item]) iter.Seq[string] { ... }

seen := map[string]struct{}{}
for id := range Rendered(Ready(items)) {
	if _, dup := seen[id]; dup {
		continue
	}
	seen[id] = struct{}{}
	fmt.Println(id)
}
```

Iterators are lazy (no intermediate slice allocation), composable, and
don't ask the reader to learn a library idiom.

Bridges to and from slices:

```go
// Slice → iter.Seq
for v := range slices.Values(xs) { ... }
for i, v := range slices.All(xs) { ... }

// iter.Seq → slice
materialized := slices.Collect(myIter)

// iter.Seq2 → map
m := maps.Collect(kvIter)
```

Sorted output:

```go
for v := range slices.Sorted(myIter) { ... }
```

### Pull-style consumption

When you want to advance an iterator manually (parser, state machine):

```go
next, stop := iter.Pull(myIter)
defer stop()

first, ok := next()
// ... decide based on first
```

Use `iter.Pull2` for `Seq2[K,V]`.

## Layer 3: plain `for` loop

Often the right answer. Two lines of range loop beat three lines of iter

- closure ceremony for one-off transformations.

```go
var evens []int
for _, n := range nums {
	if n%2 == 0 {
		evens = append(evens, n)
	}
}
```

If this only happens once, stop here. Don't extract a helper.

## Layer 4: `samber/lo` as last resort

[`samber/lo`](https://github.com/samber/lo) still has a place, but a
smaller one than before Go 1.21+. Reach for it **only** when:

- stdlib + iter don't cleanly provide the helper: `lo.GroupBy`,
  `lo.Partition`, `lo.SliceToMap`, `lo.Keyify`, `lo.Associate`.
- Your codebase already imports it and consistency matters.
- You need functional primitives `lo.Ternary` (beware: both branches
  evaluate), `lo.CoalesceOrEmpty`.

Don't reach for lo when:

- stdlib `slices`/`maps`/`cmp` has it. Always check first.
- You can express it as an `iter.Seq` chain that reads more clearly.
- You're on a hot path (`lo` closures allocate; range loops don't).
- The team hasn't adopted lo — one-off use introduces a new dep.

## Rule of thumb

| Need                                       | Go to                              |
| ------------------------------------------ | ---------------------------------- |
| Contains, Sort, Delete, Equal, Max/Min     | stdlib `slices` / `maps` / `cmp`   |
| Lazy composition, filter+map+uniq pipeline | `iter.Seq` function + range loop   |
| One-off transformation, 5-item slice       | plain `for` loop                   |
| `GroupBy`, `Partition`, `SliceToMap`       | samber/lo                          |
| `Map`/`Filter`/`Chunk` on a large slice    | **iter.Seq or plain loop, not lo** |

## Anti-patterns

- **`lo.Contains` when `slices.Contains` exists.** Use stdlib.
- **`lo.Map` + `lo.Filter` + `lo.Uniq` on a 5-item slice.** Write the loop.
- **`lo.Must(json.Unmarshal(...))` in production code.** Panics are a
  bad user experience.
- **Mixing `lo` helpers and stdlib `slices` helpers in one file.** Pick one.
- **`lo.Ternary(cond, a, b)` for side-effecting branches.** Both sides
  evaluate; use `if`/`else`.
- **Building a slice just to iterate and discard.** Return `iter.Seq`
  instead.
- **Exposing `[]T` from an API when the caller only iterates.** Return
  `iter.Seq[T]` — the caller picks materialization.
