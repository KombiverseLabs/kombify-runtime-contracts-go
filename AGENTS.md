# Agent Instructions

This repository is the public, dependency-light Go contract module for the
Kombify runtime boundary. Read the canonical workspace
`PLATFORM-STRATEGY.md`, `PLATFORM-ARCHITECTURE-TARGET.md`,
`PUBLIC-BETA-PLAN.md`, and repository manifest before changing a contract.

## Boundaries

- Keep exactly four public contract surfaces: `runtimelease`,
  `providerexecutor/v1beta1`, `runtimeinventory`, and `stackaction`.
- Do not add provider SDKs, credentials, connection-profile implementations,
  adapter registries, databases, durable product state, billing/product policy,
  or hosted-service clients.
- Do not add aliases or wrappers for `vmlease`, `runtimeaction`,
  `serverruntime`, or a previous provider-executor wire.
- `runtimeinventory` is generated from the Techstack inventory schema.
- `stackaction` is generated from the canonical StackKits CUE contract. Edit
  the CUE source and generator in StackKits, then copy the verified generation
  bundle; never hand-edit the generated Go projection.
- Public source, tests, docs, and fixtures must contain no private dependency,
  URL, secret, PII, or production data.

## Commands

```bash
mise run check
mise run test:affected -- runtimelease
mise run test:affected -- providerexecutor/v1beta1
mise run test:affected -- runtimeinventory
mise run test:affected -- stackaction
```

Use Beads for repo-local execution work, ROADMAP.md for milestone scope, and
Linear for cross-repository priority or blockers. Completed work must be
committed and pushed from an isolated branch.
