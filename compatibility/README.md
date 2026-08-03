# Consumer compatibility evidence

This directory defines the public, dependency-light evidence contract for a
consumer of `kombify-runtime-contracts-go`. It does not add a fifth Go package.

Each consumer keeps one lock matching `consumer-lock.schema.json`. The lock
binds the exact module version and Go module zip checksum (`go.sum`'s non-
`/go.mod` `h1:` line), lists only the packages the consumer actually imports,
and hashes every consumer-owned compatibility fixture. A module bump therefore
fails until the consumer deliberately refreshes and exercises its fixtures.

`corpus/manifest.json` indexes reusable, secret-free seeds for closed-JSON and
digest fuzz tests. The root compatibility test verifies every seed digest and
strictly decodes all four public wires. Consumers may copy a relevant seed,
but their lock must hash the copied fixture; a path into another checkout is
not evidence.

See [Compatibility policy](../docs/COMPATIBILITY.md) for additive and
coordinated-breaking change requirements.
