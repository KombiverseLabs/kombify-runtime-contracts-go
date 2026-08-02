---
title: kombify-runtime-contracts-go Roadmap
last_verified: 2026-08-03
roadmap_standard: kombify-roadmap@v1
generator: kombify-roadmap-sync@v1
track: v0-expansion
linear_project: Development
---

# kombify-runtime-contracts-go Roadmap

## Current Focus

- **Target:** v0.2.0 - Compatibility Evidence
- **Outcome:** Every active consumer proves exact fixture compatibility with
  the four public runtime-contract packages before contract drift can merge.
- **Exit gate:** Consumer fixtures bind the exact contract source, additive and
  breaking-change evidence is explicit, and reusable fuzz corpora cover closed
  JSON and digest validation.
- **Blocking bugs:** Beads label `blocks:v0.2.0`

## Expansion Track

| Version | Stage | State | Outcome |
| --- | --- | --- | --- |
| v0.1.0 | Runtime Contract Authority Baseline | done | Four contract packages, generation provenance, public boundary checks, and coordinated consumer compilation are proven. |
| v0.1.4 | Release Identity Integrity | done | Source version, requested Delivery version, and immutable module tag are one fail-closed identity. |
| v0.2.0 | Compatibility Evidence | current | Cross-repository fixture verification and breaking-change evidence are automated for every consumer. |
| v0.3.0 | Release Discipline | planned | Release notes, source attestations, and compatibility policy are exercised across multiple upgrades. |
| v0.4.0 | API Stability Candidate | planned | Public wire and Go evolution rules are proven across the active consumer set before any v1 commitment. |

## v0.1.0 - Runtime Contract Authority Baseline

**Scope**

- [x] Publish the `runtimelease` contract without product state or private transport dependencies.
- [x] Publish only `providerexecutor/v1beta1` for provider control.
- [x] Publish the secret-free `runtimeinventory` read model and Golden JSON.
- [x] Ingest `stackaction` only from the deterministic StackKits CUE generation bundle.
- [x] Add public dependency, provider-SDK, secret-field, and retired-path boundary checks.
- [x] Prove Techstack and bounded conformance consumers compile against one exact module revision.
- [x] Remove the transitional contract packages from `kombify-go-common` in the coordinated cutover.
- [x] Supersede the unpublished `v0.1.0` baseline with the historical `v0.1.3` publication.

**Exit gate**

- [x] `mise run check` passes from a clean checkout.
- [x] Golden JSON and StackAction generation provenance are deterministic.
- [x] No private Kombify module, provider SDK, credential type, adapter registry, product state, or legacy alias is present.
- [x] Every coordinated consumer uses the exact tagged contract version.
- [x] No open P0/P1 Beads bugs with `blocks:v0.1.0`.

## v0.1.4 - Release Identity Integrity

**Scope**

- [x] Correct the committed source version after the immutable `v0.1.3` historical release.
- [x] Bind Delivery Plan version derivation to the exact committed source version.
- [x] Reject release versions, tags, plan identity, or existing release
  metadata that differ from the exact source-derived Delivery Plan.
- [x] Keep Release Please version files aligned with its release proposal.
- [x] Publish `v0.1.4` only through an explicitly approved Delivery run from exact `main`.

**Exit gate**

- [x] Unit and workflow contract tests cover source, version, tag, plan-digest,
  and idempotent release-metadata mismatches.
- [x] The dependency-light four-package public boundary remains unchanged.
- [x] `v0.1.4` resolves to exact approved `main` and its GitHub release is verified.
- [x] No open P0/P1 Beads bugs with `blocks:v0.1.4`.

## v0.2.0 - Compatibility Evidence

**Scope**

- [x] Generate `runtimeinventory` from the Techstack-owned machine schema and
  bind the copied projection to exact source/output digests.
- [ ] Verify consumer fixtures against exact contract source hashes in CI.
- [ ] Define additive-change and coordinated breaking-change evidence.
- [ ] Add reusable fuzz corpora for closed JSON and digest validation.

**Exit gate**

- [ ] Every active consumer proves fixture compatibility in its own repository.
- [x] Techstack proves exact schema, generated Go, published module, and OpenAPI
  structural parity without a parallel field map.
- [ ] Contract drift fails before a consumer merge.
- [ ] No open P0/P1 Beads bugs with `blocks:v0.2.0`.

## v0.3.0 - Release Discipline

**Scope**

- [x] Bind plan compilation and release execution to one exact centrally
  authorized Delivery compiler revision.
- [x] Correlate the authoritative release workflow by immutable release ID and
  plan digest instead of source SHA alone.
- [ ] Exercise release notes and source attestations across multiple upgrades.
- [ ] Prove coordinated rollback to the preceding compatible contract set.

**Exit gate**

- [x] Manual or stale dispatches fail before compiler execution when their
  compiler ref is missing, malformed, or differs from the organization pin.
- [ ] At least two cross-repository upgrades complete without compatibility aliases.

## v0.4.0 - API Stability Candidate

**Scope**

- [ ] Exercise the proposed public Go and JSON evolution rules against every
  active consumer without compatibility wrappers.
- [ ] Record the evidence required to freeze the four-package surface for a
  future v1 decision.

**Exit gate**

- [ ] A v1 proposal can identify stable guarantees and remaining exceptions
  from repeated upgrade evidence rather than repository-local assumptions.
- [ ] No open P0/P1 Beads bugs with `blocks:v0.4.0`.

<!-- BEGIN GENERATED: open-issues kombify-roadmap-sync -->
## Open Issues

_Generated from Beads open statuses owned by `kombify-runtime-contracts-go`; milestone sections use
`milestone:*` / `blocks:*` labels, unmapped repo issues are listed separately, and
shared-store ownership requires an exact `repo:kombify-runtime-contracts-go` or repo-short label — do not edit;
refresh via `mise run roadmap:update -- -Repo kombify-runtime-contracts-go`. Source: `bd list`, 2026-08-02._

_Unowned shared-store issues are omitted; run the workspace ownership audit before claiming completeness._

### M1 · v0.2.0 — Compatibility Evidence (0 open)
- none

### M2 · v0.3.0 — Release Discipline (1 open)
- `platform-iqzf2` Bind public runtime release adapter to the central compiler pin and exact dispatch (P1, open · blocking)

### M3 · v0.4.0 — API Stability Candidate (0 open)
- none

### Unmapped Beads (0 open)
- none

**Total open:** 1
<!-- END GENERATED: open-issues kombify-roadmap-sync -->

## V1 Definition

- **State:** Uncommitted.
- **Known prerequisites:** Repeated consumer upgrades, stable generation
  authorities, and proven compatibility/rollback evidence.
- **Open questions:** Which additive evolution rules can be frozen for a public
  `v1` without constraining the owning product schemas prematurely.

## Later

- Fuzz and multi-language fixture projections after the Go consumer cutover is stable.

## Not Planned

- Provider adapters, SDKs, credentials, connection profiles, or adapter registry.
- Techstack product policy, durable runtime state, inventory collection, or provider ledger.
- StackKits lifecycle implementation or an independently authored action schema.
- Compatibility aliases for retired runtime packages.
