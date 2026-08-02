---
title: kombify-runtime-contracts-go Roadmap
last_verified: 2026-08-02
roadmap_standard: kombify-roadmap@v1
generator: kombify-roadmap-sync@v1
track: v0-expansion
linear_project: Development
---

# kombify-runtime-contracts-go Roadmap

## Current Focus

- **Target:** v0.1.4 - Release Identity Integrity
- **Outcome:** Runtime consumers share one dependency-light public Go module
  for runtime leases, provider execution, inventory, and StackKits actions.
- **Exit gate:** Golden fixtures and generation provenance pass, Techstack and
  bounded conformance consumers compile against the exact module revision, and
  no retired compatibility path remains in the coordinated cutover.
- **Blocking bugs:** Beads label `blocks:v0.1.4`

## Expansion Track

| Version | Stage | State | Outcome |
| --- | --- | --- | --- |
| v0.1.0 | Runtime Contract Authority Baseline | superseded | Four contract packages, generation provenance, public boundary checks, and coordinated consumer compilation are proven. |
| v0.1.4 | Release Identity Integrity | current | Source version, requested Delivery version, and immutable module tag are one fail-closed identity. |
| v0.2.0 | Compatibility Evidence | planned | Cross-repository fixture verification and breaking-change evidence are automated for every consumer. |
| v0.3.0 | Release Discipline | planned | Release notes, source attestations, and compatibility policy are exercised across multiple upgrades. |

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
- [ ] No open P0/P1 Beads issues with `blocks:v0.1.0`.

## v0.1.4 - Release Identity Integrity

**Scope**

- [x] Correct the committed source version after the immutable `v0.1.3` historical release.
- [x] Bind Delivery Plan version derivation to the exact committed source version.
- [x] Reject release versions, tags, plan identity, or existing release
  metadata that differ from the exact source-derived Delivery Plan.
- [x] Keep Release Please version files aligned with its release proposal.
- [ ] Publish `v0.1.4` only through an explicitly approved Delivery run from exact `main`.

**Exit gate**

- [x] Unit and workflow contract tests cover source, version, tag, plan-digest,
  and idempotent release-metadata mismatches.
- [x] The dependency-light four-package public boundary remains unchanged.
- [ ] `v0.1.4` resolves to exact approved `main` and its GitHub release is verified.
- [ ] No open P0/P1 Beads bugs with `blocks:v0.1.4`.

## v0.2.0 - Compatibility Evidence

**Scope**

- [ ] Verify consumer fixtures against exact contract source hashes in CI.
- [ ] Define additive-change and coordinated breaking-change evidence.
- [ ] Add reusable fuzz corpora for closed JSON and digest validation.

**Exit gate**

- [ ] Every active consumer proves fixture compatibility in its own repository.
- [ ] Contract drift fails before a consumer merge.

## v0.3.0 - Release Discipline

**Scope**

- [ ] Exercise release notes and source attestations across multiple upgrades.
- [ ] Prove coordinated rollback to the preceding compatible contract set.

**Exit gate**

- [ ] At least two cross-repository upgrades complete without compatibility aliases.

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
