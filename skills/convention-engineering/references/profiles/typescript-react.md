# TypeScript/React Convention Profile

Conventions for TypeScript/React frontend repositories.

## Build System: Taskfile.dev + Bun/Vite

Taskfile wraps bun and vite commands:
```yaml
tasks:
  client:dev:
    desc: Start client dev server
    dir: packages/client
    cmds: [bun run dev]
  client:gen:
    desc: Generate client GraphQL SDK
    cmds: [bun run codegen]
  build:web:
    desc: Build React frontend
    dir: web
    cmds: [bun run build]
```

### Vite Configuration
- Manual chunk splitting for vendor libraries (React, MUI/Arco, Chart.js)
- API proxy for dev server (`/api` -> backend)
- Source maps for production debugging

## Type Safety

### Non-Negotiable
- TypeScript strict mode always on (`"strict": true`)
- No `any` types without justification
- **Forbidden patterns**: `as unknown as`, `@ts-ignore`, phantom interfaces for casting
- Type errors = code generation checkpoint, not a casting opportunity

### Schema-First Development (GraphQL)
- Introspect GraphQL schema BEFORE writing queries (never assume field names)
- Code generation order: `task server:gen` -> restart server -> `task client:gen`
- Never edit generated files (marked DO NOT EDIT)

## State Management: Jotai

- Atoms in `models/*/atom.ts`
- RxJS for complex async flows
- No Redux (too much boilerplate for this scale)

## Component Patterns

- Use component library Layout/Space/Grid over custom div+flexbox
- Optimistic UI updates for user interactions
- Batch loading for list queries (avoid N+1)

## i18n

- Use `transRaw()` for all user-facing strings (not `trans()`)
- Library: `@ies/starling_intl`
- Never hardcode user-visible text

## Linting: ESLint + Prettier

- ESLint for logic rules (react-hooks, react-refresh, typescript-eslint)
- Prettier for formatting (80-char line limit)
- Enforced via lint-staged + Husky pre-commit
- Commitlint for conventional commits

## Testing: Vitest

- Framework: Vitest (Vite-native, faster than Jest)
- Assertions: @testing-library/react, @testing-library/jest-dom
- Coverage target: 64%+ baseline
- Environment: jsdom

## Quality Gates

Adapt task names to your project. Common patterns:

```bash
npm run lint            # ESLint + Prettier check
npm test                # Vitest unit tests
npm run build           # Vite production build (catches import errors)
npx tsc --noEmit        # TypeScript type check
```

## File Organization

```
packages/client/src/
├── pages/              # Route-level components
├── common/             # Shared components
├── hooks/              # Custom React hooks
├── models/             # Jotai atoms and state logic
├── constants/          # Enums, config constants
└── frame/              # App shell, navigation
```

Files should not exceed 300 lines. Split large components.
