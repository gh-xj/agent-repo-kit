# Core Belief

This repo exists to make future agent work more legible, reliable, and
compounding.

`agent-repo-kit` is not just a bundle of templates. It is a belief that
software work improves when the repo itself carries the context, constraints,
and feedback loops that agents need to operate well.

## Beliefs

1. Agents are first-class operators.
   Design repo structure, commands, docs, and verification so an agent can
   safely continue work without private memory or hidden context.

2. Hidden state does not exist.
   Knowledge in chat, issue trackers, Slack, or a person's head is invisible
   to future agents. If it matters, make it a versioned artifact.

3. Simplicity means fewer canonical concepts, not fewer files.
   A system with a small durable model and clear derived views is simpler than
   a compact system that overloads one concept with many meanings.

4. Derived views beat duplicated state.
   Store canonical facts once. Let lists, boards, dashboards, and status
   summaries be recomputed from those facts.

5. Durable models must outlive current tools.
   Design persistent state around stable domain concepts, not around the shape
   of today's CLI, SaaS product, prompt, or agent runtime.

6. Human judgment should become reusable constraints.
   When review catches a recurring mistake, encode the lesson as a rule, test,
   lint, template, or documented invariant.

7. Evidence beats assertion.
   A claim that work is done should point to the artifact, command, test, or
   observation that makes it true.

8. Parallel agents require explicit coordination.
   Shared files, IDs, leases, and work ownership need clear coordination
   primitives. Hope is not a concurrency strategy.

9. Local-first is the default until a networked system earns its cost.
   Local files are inspectable, versionable, scriptable, and resilient. Add
   hosted state only when it solves a real coordination problem.

10. Compatibility comes from clean primitives, not premature fields.
    Future-proofing means preserving room to evolve. Do not add concepts just
    because they may be useful later.

## Decision Rules

When tradeoffs conflict:

1. Prefer canonical state over convenience state.
2. Prefer derived views over stored duplicates.
3. Prefer explicit invariants over prose-only guidance.
4. Prefer small durable concepts over feature mimicry.
5. Prefer migration docs over compatibility baggage when replacing a bad
   model.
6. Prefer boring, inspectable storage over clever machinery.
7. Prefer a missing feature over a premature abstraction.

## Anti-Beliefs

- Do not copy SaaS products wholesale.
- Do not add fields only because they might be useful later.
- Do not preserve legacy abstractions when the goal is replacement.
- Do not let agent docs become encyclopedias.
- Do not treat unverified claims as completed work.
- Do not let a view become a second source of truth.
- Do not use "future-proof" to justify speculative complexity.

