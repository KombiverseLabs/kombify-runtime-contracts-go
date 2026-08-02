# runtimelease

`runtimelease` is the secret-free contract for the Runtime Lease Authority.
Techstack owns the durable lease state; this package only defines the wire and
local validation behavior.

The authority owns four dimensions:

- lease identity and validity window
- caller-owned desired state (`running`, `stopped`, or `absent`)
- cancellation
- a monotonic, non-zero `revision` used for compare-and-swap mutations

Provider execution, provider observations, cleanup, server lifecycle, Guard
health/connection/inventory, billing, catalog policy, and access custody are
separate authorities and do not appear in this contract.

## Identity model

A `server_id` identifies one durable RuntimeServer. Replacing infrastructure
does not replace that identity. A `resource_generation_id` identifies exactly
one replaceable infrastructure generation and must be a non-nil canonical
lowercase UUID. Lease validation and enrollment always fence both identities
plus the exact lease revision.

Tenant ownership and the responsible owner are explicit `tenant_id` and
`owner_id` fields. Provider names, adapter names, native resource handles,
simulation identifiers, VM identifiers, provisioning specifications, observed
state, arbitrary metadata, credentials, and access material are intentionally
absent.

## Wire contract

There is no `vmlease` compatibility package. Validation requests carry `tenant_id`,
`owner_id`, `lease_id`, `lease_revision`, `server_id`, and
`resource_generation_id`.
Validation failures use the closed reason set:

- `lease_cancelled`
- `lease_not_yet_valid`
- `lease_expired`
- `binding_mismatch`
- `revision_mismatch`
- `invalid_lease`

Enrollment requests and acknowledgements repeat the same exact identity fence.
Their idempotency key belongs to the enrollment authority and is not forwarded
as a provider idempotency key.

## Snapshots

Snapshots sign the complete lease projection with an authority-owned Ed25519
private key. The wire fixes `runtimelease-snapshot/v1`, `Ed25519`, and a signed
`key_id`; consumers receive only a rotation-capable public-key set and cannot
mint leases. `expires_at` is the normal validity boundary and may be at most
five minutes after issuance. Optional
`grace_until` is chosen and signed by the authority, is capped at five minutes,
and is accepted only by `VerifySnapshotFallback` when `FallbackEligible`
classifies the live authority failure as an outage. A caller cannot extend
grace, and an explicit invalid result from the authority must never fall back
to a cached snapshot.

Consumers verify every returned snapshot against a configured public-key
resolver, issuer, audience, owner, lease revision, server, and resource
generation. Authenticated HTTP transport and authority routing belong to the
owning runtime, not this module.

## Packaging

The package is public and has no provider SDK, database, durable claim, adapter
configuration, private module, or product-policy dependency.
