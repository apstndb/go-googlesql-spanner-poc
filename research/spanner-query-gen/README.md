# spanner-query-gen Research

This directory keeps non-normative design notes and review exchanges for
`cmd/spanner-query-gen`.

Current command documentation is under
[`cmd/spanner-query-gen/`](../../cmd/spanner-query-gen/). Generated JSON schemas
are under [`schemas/`](../../schemas/). Those files, plus tests, are the source
of truth for the current v1alpha behavior.

## Files

- [`PLAN_CONTRACT_CANDIDATES.md`](PLAN_CONTRACT_CANDIDATES.md): candidate
  query-plan contracts and optimization-practice notes. The implemented
  contract surface remains documented under
  [`cmd/spanner-query-gen/PLAN_CONTRACTS.md`](../../cmd/spanner-query-gen/PLAN_CONTRACTS.md).
- [`reviews/`](reviews/): archived review prompts, review responses, and design
  discussion notes used while iterating on the v1alpha configuration, code
  generation, external dataset, and plan-contract design.
