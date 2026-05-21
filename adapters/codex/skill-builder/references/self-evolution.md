# Self-Evolution Checklist

Use this reference when a skill is mature enough to need decay protection: it
has real walkthroughs, external dependencies, recurring drift, or lessons that
currently have nowhere durable to land.

Do not add this scaffolding to throwaway or brand-new skills. Premature
maintenance files become ceremony.

## Five-Item Audit

For a target skill at `<skill-path>/`, check each item with PASS / FAIL and a
one-line justification.

1. **Append-only lesson log.** Does the skill have `maintenance.md` or
   `MAINTENANCE.md` with a Lessons or Changelog section?
2. **Failure registry when warranted.** If the skill depends on external CLIs,
   APIs, browser flows, or services, does it have `known-failures.md` or an
   equivalent workaround registry? Pure prose skills can skip this.
3. **Feedback trigger.** Does the consuming workflow include a closeout prompt
   or maintenance loop that asks whether a lesson should be captured?
4. **Explicit anti-scope.** Does `SKILL.md` include "When not to use",
   "Boundaries", or another clear anti-trigger section?
5. **Falsifiers for absolutes.** Mandatory rules such as "always" or "never"
   should include a condition that would prove the rule wrong or justify
   revising it.

## Patch Policy

- For user-requested refactors, apply the smallest patch that fixes the failed
  audit item.
- For unsolicited audits, report findings first and patch only after the user
  agrees.
- If an item is intentionally skipped, record the reason and falsifier in the
  skill's maintenance file instead of silently omitting it.
- Keep the lesson log append-only. Reverting a lesson is a new entry.

## Minimum Maintenance File

```markdown
# <skill-name> Maintenance

## When This File Changes

- A real run changes a rule, trigger, workflow, reference, or script.
- An external dependency drifts and the skill needs a new workaround or rule.
- A self-evolution audit item is intentionally skipped.

## Lessons

### YYYY-MM-DD - short title

**Trigger**: what run, failure, or user correction surfaced this.
**Change**: file plus one-line summary.
**Expected effect**: what improves.
**Falsifier**: what would prove the change wrong.
```

## Known Failures File

Only add this when the skill has runtime dependencies or recurring operational
failures.

```markdown
# <skill-name> Known Failures

## <symptom> (YYYY-MM-DD)

**Status**: warning-only | recoverable | blocking | fixed in <commit>
**Symptom**: what the operator sees.
**Workaround**: what to do now.
**Permanent fix candidate**: what would close it at the source.
**Cross-ref**: maintenance entry or work item.
```

## Healthy Result

The target skill remains small at the front door, has one durable place for
lessons, and has a trigger that makes future agents write lessons when real use
changes the workflow.
