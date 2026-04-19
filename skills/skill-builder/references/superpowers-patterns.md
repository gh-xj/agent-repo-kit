# Superpowers Patterns for Skill Engineering

Patterns extracted from the superpowers skill library that improve skill quality. Apply these when creating or updating skills.

## 1. Skill Testing Protocol (RED-GREEN-REFACTOR)

Treat skill changes like code changes — test before shipping.

### Workflow

1. **RED**: Capture baseline agent behavior WITHOUT the new rule. Run a subagent with the skill loaded but the new rule absent. Record what it does wrong.
2. **GREEN**: Add the minimal rule that corrects the behavior. Re-run the same scenario. Verify compliance.
3. **REFACTOR**: Plug rationalizations. Run adversarial scenarios where the agent might skip the rule. Add explicit counters for each rationalization discovered.

### When to Apply

- Adding new guardrails or workflow steps to a skill
- Changing routing logic or decision gates
- Updating severity calibration or assessment labels

### Rationalization Table

For each non-negotiable rule, add an "Excuse/Reality" pair:

```markdown
## Rationalization Prevention

| Excuse | Reality |
|--------|---------|
| "I'll reconcile artifacts later" | Artifacts are lost across dispatch cycles |
| "This case is simple enough to skip triage" | Simple cases have the most unexamined assumptions |
| "The fix is obvious, no need for evidence" | Obvious fixes have the highest revert rate |
```

---

## 2. Persuasion Engineering

Use behavioral principles to make skills stick. These patterns reduce decision fatigue and prevent agent drift.

### Authority Language (HARD-GATE)

For non-negotiable rules, use explicit authority markers:

```markdown
<HARD-GATE>
Do NOT proceed past this point without [condition].
This applies to EVERY case regardless of perceived simplicity.
</HARD-GATE>
```

Reserve for truly critical gates — overuse dilutes impact.

### Scarcity / Temporal Gates

Time-bound requirements prevent procrastination:

```markdown
IMMEDIATELY after worker sidecars complete, reconcile artifacts into case.md
before creating the next commit.
```

Use "IMMEDIATELY after X", "Before proceeding to Y" for ordering-critical steps.

### Commitment Announcements

Force the agent to commit to a decision before acting:

```markdown
Before dispatching workers, announce in case.md:
- Risk classification: [low | standard | high]
- Worker routing: [list of workers]
- Budget: [worker count]

This forces review of the routing decision before execution.
```

### Social Proof (Norm-Setting)

Establish universal patterns as norms:

```markdown
Every tracked request becomes a case. Always.
Worktrees without case.md ownership = abandoned state. Always.
```

---

## 3. Subagent Prompt Engineering

When skills dispatch subagents, structure prompts for reliability.

### Implementer Prompt Pattern

```markdown
## Context
[Paste full case scope inline — never pass a file path for the agent to read]

## Before You Begin
Ask up to 5 clarifying questions before starting work. Wait for answers.

## Self-Review Checklist (before reporting)
- [ ] All requirements addressed?
- [ ] No scope creep beyond what was asked?
- [ ] Tests written and passing?
- [ ] Existing code patterns followed?
```

### Reviewer Prompt Pattern (Spec Compliance)

```markdown
Do NOT trust the implementer's report. Verify actual code vs. requirements.

Check for:
1. Missing requirements (claimed done but not implemented)
2. Extra features not in spec (scope creep)
3. Misunderstood requirements (implemented wrong thing)

Verdict: COMPLIANT or NON-COMPLIANT (binary, no "close enough")
```

### Two-Stage Review Order

Spec compliance BEFORE code quality. Never skip either. Wrong order = invalid result.

```
Implementer -> Spec Compliance Reviewer (pass/fail gate)
                    |
                    v (only if COMPLIANT)
              Code Quality Reviewer
```

### Parallel Read Mandate

When dispatching agents that need to analyze multiple files:

```markdown
Read ALL source files in parallel before analyzing.
Do NOT read files sequentially — this wastes time and misses cross-file patterns.
```

---

## 4. Evidence-Before-Claims

No success claims without evidence. From the verification-before-completion pattern.

### Iron Law

> Never claim work is complete, tests pass, or a fix works without running the verification command and reading the full output in this session.

### Red-Flag Language

These words signal guessing instead of verifying:

| Red Flag | Replace With |
|----------|-------------|
| "should work" | Run it and show output |
| "probably fixed" | Run test and paste result |
| "seems to pass" | Show exact pass/fail output |
| "I believe" | Provide evidence |

### Skill Application

In skills that produce verdicts or status updates:

```markdown
## Guardrail: Evidence Gate

Before setting any status to `resolved` or `verified`:
1. Run the verification command specified in the case
2. Read and quote the relevant output
3. Status claim must include artifact path or command output
```

---

## 5. Architectural Escalation

From systematic-debugging: if the same component fails 3+ times, question the architecture.

```markdown
## Escalation Rule

If the same file or function has been modified 3+ times in a single case:
1. Stop implementing fixes
2. Flag "Architectural concern" in case.md
3. Propose structural change before attempting fix #4
```

---

## 6. Condition-Based Waiting

For skills involving async operations or polling, replace arbitrary timeouts:

```markdown
## Anti-Pattern: Arbitrary Timeouts
BAD:  time.Sleep(2 * time.Second)
GOOD: WaitForCondition(ctx, condition, "message appeared", 10*time.Second)

## Pattern
func WaitForCondition(ctx context.Context, check func() bool, desc string, timeout time.Duration) error
- Poll every 100ms
- Clear error message on timeout: "Timed out waiting for: {desc}"
- Respect context cancellation
```

---

## 7. Skill Review Checklist (Extended)

Add these checks to the existing Quality Checklist:

- [ ] **Non-negotiable rules use HARD-GATE markers** (authority language)
- [ ] **Ordering-critical steps use temporal gates** ("IMMEDIATELY after X")
- [ ] **Subagent prompts paste context inline** (never pass file paths for agents to read)
- [ ] **Review workflows are two-stage** (spec compliance then code quality)
- [ ] **Verdict fields are binary** (pass/fail, not prose)
- [ ] **Status updates require evidence** (artifact path or command output)
- [ ] **Rationalization table exists** for rules agents commonly skip
- [ ] **Parallel read mandate** included in agent dispatch prompts
