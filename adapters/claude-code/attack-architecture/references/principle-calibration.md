# Principle Calibration

Principles help reviewers aim the attack. They do not prove findings.

Use a principle only when it sharpens a concrete accusation:

```
code evidence -> architectural pressure -> principle tag
```

Do not use the reverse shape:

```
principle name -> vague accusation -> no concrete failure mode
```

## Allowed Tags

Use at most three tags per finding:

- `DRY`
- `KISS`
- `YAGNI`
- `SoC`
- `SRP`
- `OCP`
- `LSP`
- `ISP`
- `DIP`
- `Abstraction`
- `Encapsulation`
- `Cohesion`
- `Coupling`
- `Modularity`
- `DesignForFailure`
- `Observability`
- `BoundedContext`
- `ExplicitDependency`
- `OperationalExcellence`
- `Security`
- `Reliability`
- `PerformanceEfficiency`
- `CostOptimization`
- `Sustainability`

## Principle Map

| Principle family                              | Attack use                                                                                                        |
| --------------------------------------------- | ----------------------------------------------------------------------------------------------------------------- |
| DRY                                           | Duplicate knowledge, copy-paste families, parallel hierarchies, or two sources of truth.                          |
| KISS / YAGNI                                  | Speculative abstraction, unused flexibility, clever indirection, premature async/cache/batch machinery.            |
| SoC / SRP / cohesion / coupling / modularity  | Mixed responsibilities, boundary leaks, god modules, features scattered across unrelated files.                    |
| Abstraction / encapsulation / OCP / LSP / ISP | Leaky types, subclass/interface traps, caller-visible internals, contracts that force clients to know too much.    |
| Dependency inversion / explicit dependencies  | Hidden globals, lower layers importing upper layers, framework/vendor coupling where a domain boundary should sit. |
| Design for failure / observability            | Missing propagation, missing context, retry/fallback behavior that hides failures, unobservable critical paths.    |
| Bounded contexts                              | Cross-context data ownership confusion, vocabulary drift, DTO/domain/persistence contracts that bleed together.    |
| TDD / CI/CD / documentation                   | Only attack when absence or drift makes the design unprovable, unchangeable, or dependent on tribal knowledge.     |
| Well-Architected-style pillars                | For cloud/platform workloads, calibrate against operational excellence, security, reliability, performance efficiency, cost optimization, and sustainability. Dedicated security review remains out of scope. |

## Anti-Patterns

- "Violates SOLID" as a title or conclusion.
- "Not DRY" without naming the duplicate knowledge and the future drift.
- "Not scalable" without a cost shape, ownership boundary, or hot path.
- "Needs observability" without a concrete failure path that cannot be
  diagnosed.
- Treating the cloud pillars as a generic checklist for every repo.
- Turning a security concern into a finding when it needs a dedicated security
  review instead of architecture pressure analysis.

## Lens Placement

Do not add a cloud or SOLID super-lens just because a principle applies. Route
the pressure to the existing lens that can produce the sharpest evidence:

- L1 Overengineering: KISS, YAGNI, disproportionate cloud/platform machinery.
- L2 Data-model / contract: illegal states, bounded-context bleed, leaky DTOs.
- L3 Coupling & boundaries: SoC, SRP, DIP, explicit dependencies.
- L4 Silent failures: design for failure, observability, reliability.
- L5 Evolvability: DRY drift, shotgun surgery, process as architecture.
- L6 Concurrency: lifecycle ownership and state discipline.
- L7 Performance: performance efficiency, cost optimization, sustainability.
