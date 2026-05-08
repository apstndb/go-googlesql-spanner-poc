# Research

This directory contains non-normative research notes, upstream feedback drafts,
and review archives.

The public user-facing documentation remains in the repository root and under
`cmd/`. Tool usage documentation stays next to the tool under `tools/`.
Research files can be longer, more exploratory, and tied to a specific local
probe environment.

## Areas

- [`spanner-query-plan-shape/`](spanner-query-plan-shape/): Spanner Omni query
  plan shape observations, optimizer-version matrices, and feedback prepared
  for Spanner Unofficial Hacks.
- [`spanner-query-gen/`](spanner-query-gen/): design-review exchanges and
  non-normative notes used while shaping `cmd/spanner-query-gen`.

Treat files in this directory as evidence and background. When they disagree
with generated schemas, tests, or command documentation, the implementation
surface should be checked directly before changing behavior.
