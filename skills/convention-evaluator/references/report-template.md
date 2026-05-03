# Report Template

Recommended layout for `docs/reviews/YYYY-MM-DD_<topic>_evaluation.md`.

```markdown
# Convention Evaluation — `<repo>` (<YYYY-MM-DD>)

**Status:** PASS | FAIL
**High-risk:** yes | no — <one-line justification>
**Thresholds applied:** legibility >= N, enforceability >= N, verification >= N

## Scores

| Dimension      | Score (0-4) | Threshold | Verdict |
| -------------- | ----------- | --------- | ------- |
| legibility     | N           | M         | ✓ / ✗   |
| enforceability | N           | M         | ✓ / ✗   |
| verification   | N           | M         | ✓ / ✗   |

## Findings

### legibility

<one-paragraph judgment, with file: line citations for evidence>

### enforceability

<one-paragraph judgment, with file: line citations for evidence>

### verification

<one-paragraph judgment, with file: line citations for evidence>

## What would move this from N to N+1

For each dimension that did not max out, name the smallest concrete change
that would lift the score. No vague advice — point at a file or command.

## Notes

- Optional. Use only if there is a finding that does not fit a dimension
  (e.g. an unclear opt-in in `.conventions.yaml` itself).
```

## Cite, do not paraphrase

Every finding should include a path or command output. The reader of the
report should be able to verify each claim by running the cited command or
opening the cited file.
