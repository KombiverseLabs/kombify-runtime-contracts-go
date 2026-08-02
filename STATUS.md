---
title: STATUS
last_verified: 2026-08-02
maturity: alpha
---

# Status

## Features

- The module path is `github.com/KombiverseLabs/kombify-runtime-contracts-go`.
- `runtimelease` provides the strict lease, enrollment, validation, and signed
  snapshot contract without HTTP transport or product state.
- `providerexecutor/v1beta1` provides the single provider-control wire with
  bounded envelopes, deterministic digests, append validation, and
  provider-native absence evidence requirements.
- `runtimeinventory` provides secret-free read DTOs and canonical Golden JSON.
- `stackaction` is populated only from the StackKits-generated contract bundle;
  local hand-authored wire types are not accepted.
- Root boundary tests reject private Kombify imports, provider SDKs, forbidden
  legacy package paths, and invalid Golden JSON.

## Dependencies

- Go standard library.
- `github.com/google/uuid` only, for strict canonical UUID parsing in runtime
  identity fences.
- No private repository, provider SDK, database, credential, adapter-registry,
  product-state, or hosted-service dependency.

## Tests

- Package unit and wire-compatibility tests cover lease validation and signed
  snapshots, provider command/receipt state transitions, inventory field
  safety, generated StackAction wire behavior, and Golden JSON.
- The release adapter recompiles the plan with the same immutable central
  compiler used by Delivery v2, rejects a different plan digest or identity,
  and accepts a repeated release only when all GitHub release metadata matches.
- CI runs formatting, `go vet`, race-enabled tests, build, and the boundary
  suite with a ten-minute job budget.

## Known Issues

- The immutable historical `v0.1.3` tag points at source that still declared
  `0.1.0`. Source now declares `0.1.4`, and every future Delivery release must
  bind its requested version and tag to that exact committed declaration.
- `v0.1.4` is not tagged or released by this correction change; publication
  remains an explicit Delivery action after the change lands on `main`.
