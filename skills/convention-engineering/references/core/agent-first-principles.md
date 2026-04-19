# Agent-First Principles

The invariants convention-engineering optimizes for. Every reference in this
skill should be consistent with these; when another doc conflicts, this wins.

These are **what-is-true** statements. _How_ to achieve them is the author's
call and belongs in downstream references, templates, and instances.

## What the agent sees

1. **Anything not in the repo doesn't exist.** To an agent working in-context,
   knowledge that lives in Slack, docs, chat threads, or people's heads is
   invisible. If it matters, it is a versioned artifact.

2. **Context is scarce; the map is not the encyclopedia.** `AGENTS.md` /
   `CLAUDE.md` point at deeper sources of truth. They do not contain them.
   Progressive disclosure beats a 1,000-page manual that crowds out the task.

3. **Legibility is the goal.** Code, docs, and tooling are optimized for
   future agent runs first. Human aesthetics come second. Boring technologies,
   explicit structure, and mechanically-checkable invariants all serve this.

## How judgment flows

4. **Humans steer; agents execute.** Engineering work shifts toward designing
   environments, specifying intent, and building feedback loops. Writing code
   is delegated.

5. **Encode judgment once; enforce it continuously.** When human taste catches
   a mistake, the fix is not a single review comment — it is a rule, a lint,
   a test, or a convention that applies everywhere from that day forward.

6. **When an agent fails, ask what capability is missing.** "Try harder" is
   never the answer. The fix is almost always a tool, an abstraction, a
   document, or a constraint that makes the next attempt tractable.

## How conventions land

7. **Specify invariants; leave implementation free.** A convention says what
   must be true, not how to achieve it. `exec-plans must be checked in with
decision logs` is durable. `exec-plans must have a Summary heading in
sentence case` is ritual.

8. **Mechanical enforcement > documentation > folklore.** A rule a linter can
   check is worth ten paragraphs of prose. A rule only in prose is worth ten
   minutes of tribal memory. Prefer checks over words, and words over silence.

9. **Templates, not text surgery.** When a convention scaffolds into a repo,
   it copies a template and tells the adopter what to paste where. It does
   not `awk`-patch agent contract files; that pattern has no upgrade path.

10. **Commit the convention separately from the instance.** The pattern in
    `references/templates/` and the live `.tickets/` in a repo drift for
    different reasons. Keep their commit histories distinguishable.

## How the system decays, and how we counter it

11. **Convention without enforcement becomes drift.** Every un-checked rule
    is on a slow timer. Either encode it in a lint, delete it, or accept
    that it will be ignored within a quarter.

12. **Pay debt continuously.** Technical debt is a high-interest loan. Small
    daily refactors compound better than quarterly cleanup bursts. Agents
    make the daily option cheap; take it.
