# Architecture

This repository is a pure Go contract library. It owns public type identities,
closed validation rules, deterministic digests, and compatibility fixtures. It
has no runtime process, network authority, database, provider adapter, or
credential custody.

## Ownership seams

- `runtimelease` expresses an exact lease revision and runtime generation.
  Techstack remains the validity, cancellation, desired-state, enrollment, and
  persistence authority.
- `providerexecutor/v1beta1` expresses admitted commands, unchained execution
  results, append-only receipts, resource graphs, and evidence. Techstack owns
  adapter selection, side effects, custody, CAS, and the durable ledger.
- `runtimeinventory` is a generated, secret-free read model. Techstack owns its
  schema and all collection, freshness, authorization, and access-link policy.
- `stackaction` is generated from StackKits CUE. StackKits owns lifecycle
  meaning, validation, execution, and evidence production.

The module may validate values in memory and `runtimelease` may verify signed
snapshots. It does not perform authenticated HTTP calls or persist any value.

## Generation and compatibility

Generated bundles include a generator version, source SHA-256, and output
SHA-256 values. CI validates the checked-in projection and Golden JSON. A
consumer pins one exact tagged module version; migrations are direct and never
use compatibility packages for retired import paths.
