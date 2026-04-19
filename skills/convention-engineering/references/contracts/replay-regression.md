# Replay Regression Contract

## Scenario Requirements

- scenario id and owner
- fixed replay window and fixture pack
- explicit stage chain
- user-visible assertions
- structured JSON artifact output

## Stage Chain (Reference)

`pyscript -> collector -> backend/ES -> frontend query path`

## Gate Split

- PR: minimal critical scenarios, low/no chaos.
- Nightly: full scenario matrix with chaos labels.
- Release: curated high-risk scenarios, strict thresholds.

## Failure Classes

- `new_product_logic`
- `fixture_or_time_drift`
- `infra_or_flake`
- `true_regression`

Unknown class is a hard failure.
