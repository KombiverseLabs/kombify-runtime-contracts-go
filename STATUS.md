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
- CI runs formatting, `go vet`, race-enabled tests, build, and the boundary
  suite with a ten-minute job budget.

## Known Issues

- The first `v0.1.0` tag remains intentionally unpublished until Techstack and
  the other coordinated consumers compile against this module at one exact
  revision.
