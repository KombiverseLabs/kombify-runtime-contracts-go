---
title: STATUS
last_verified: 2026-08-03
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
- `runtimeinventory` provides secret-free read DTOs and canonical Golden JSON,
  copied byte-for-byte from the Techstack generator bundle and bound to the
  canonical schema by exact source/output SHA-256 values.
- `stackaction` is populated only from the StackKits-generated contract bundle;
  local hand-authored wire types are not accepted.
- Root boundary tests reject private Kombify imports, provider SDKs, forbidden
  legacy package paths, and invalid Golden JSON.
- The public consumer-lock schema binds an exact module version and Go module
  checksum to consumer-owned fixture hashes. The shared corpus strictly
  decodes all four wires and fuzzes closed JSON plus deterministic digests.

## Dependencies

- Go standard library.
- `github.com/google/uuid` only, for strict canonical UUID parsing in runtime
  identity fences.
- No private repository, provider SDK, database, credential, adapter-registry,
  product-state, or hosted-service dependency.

## Tests

- Package unit and wire-compatibility tests cover lease validation and signed
  snapshots, provider command/receipt state transitions, Runtime Inventory
  generation lineage and field safety, generated StackAction wire behavior,
  and Golden JSON.
- The plan captures the exact organization-pinned central compiler revision;
  the release adapter rejects a missing, malformed, or changed compiler ref
  before checkout, recompiles with that exact revision, and correlates the
  authoritative workflow run by release identity plus plan digest.
- The release adapter rejects a different plan digest or identity and accepts
  a repeated release only when all GitHub release metadata matches.
- CI runs formatting, `go vet`, race-enabled tests, build, and the boundary
  suite with a ten-minute job budget.

## Known Issues

- The immutable historical `v0.1.3` tag points at source that still declared
  `0.1.0`. The corrective `v0.1.4` prerelease resolves exactly to
  `b9843afb7cc05ce0c7f408086998aaa735d197aa`; future Delivery releases bind the
  requested plan, version, tag, source, and existing release metadata.
- Techstack and the bounded Simulate conformance consumer still need to land
  their repository-local locks and exercised fixtures before the `v0.2.0`
  cross-repository exit gate is complete.
