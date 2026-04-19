# Architecture Contracts

Mechanical enforcement of architecture boundaries. Prose-only rules are not contracts -- they're suggestions. Real contracts are machine-checkable.

## Table of Contents

- [Layer Boundary Enforcement](#layer-boundary-enforcement)
- [Interface Assertion Policy (Go)](#interface-assertion-policy-go)
- [Complexity Thresholds](#complexity-thresholds)
- [String Literal Centralization](#string-literal-centralization)
- [Doc Mirror Enforcement](#doc-mirror-enforcement)

## Layer Boundary Enforcement

### Go: depguard (golangci-lint)

The gold standard. Define allowed/denied imports per package in `.golangci.yml`:

```yaml
# .golangci.yml
linters:
  settings:
    depguard:
      rules:
        handler-layer:
          list-mode: lax
          files: ["**/handler/**", "!$test"]
          deny:
            - pkg: "example.com/project/dal"
              desc: "handler must not import dal directly -- use service layer"
            - pkg: "example.com/project/operator"
              desc: "handler must not import operator directly"
        model-layer:
          list-mode: lax
          files: ["**/model/**"]
          deny:
            - pkg: "example.com/project/handler"
            - pkg: "example.com/project/service"
            - pkg: "example.com/project/operator"
            - pkg: "example.com/project/dal"
            - pkg: "example.com/project/config"
            - pkg: "example.com/project/pkg"
```

Reference architecture (5-layer):
```
handler  -> service, model, config       (never dal, operator, pkg)
service  -> operator, dal, model, config
operator -> operator/rules, model, pkg/* (never service, handler)
dal      -> model, pkg/staticdata        (never service, operator)
model    -> nothing internal             (zero deps guaranteed)
```

### TypeScript: eslint-plugin-boundaries or no-restricted-imports

```js
// eslint.config.js (flat config)
export default [
  {
    rules: {
      "no-restricted-imports": ["error", {
        patterns: [{
          group: ["../dal/*"],
          message: "Components cannot import dal directly -- use hooks/services",
        }],
      }],
    },
  },
];
```

For larger projects, use `eslint-plugin-boundaries` with element types and boundary rules.

### Python: import-linter

```ini
# .importlinter
[importlinter]
root_package = myproject

[importlinter:contract:layers]
name = Layer contract
type = layers
layers =
    api
    service
    repository
    model
```

## Interface Assertion Policy (Go)

Every interface-to-concrete pairing must have a compile-time assertion:

```go
var _ Store = (*MongoStore)(nil)
var _ Store = (*MemoryStore)(nil)
```

Update whenever interface or struct changes. This catches drift at compile time, not runtime.

## Complexity Thresholds

| Stack | Tool | Cognitive Limit | Cyclomatic Limit |
|-------|------|-----------------|------------------|
| Go | revive | 20 | 20 |
| TypeScript | eslint complexity rule | 15 | -- |
| Python | ruff C901 | 15 | -- |

## String Literal Centralization

Domain literals used across packages should be promoted to a central location:
- **Go**: Named constants in `model/` package (e.g., `model.EventFieldToolName`)
- **TypeScript**: Constants file or enum module
- **Python**: Constants module or Enum class

Preference for migration-safe aliases (`type X = string`) over strict types in Go.

## Doc Mirror Enforcement

Machine-checkable via contract checker `check-docs-governance` command:
- Define required file paths
- Define forbidden legacy paths
- Require marker sections/phrases in key docs
- Fail with explicit detail (which file, which marker missing)
