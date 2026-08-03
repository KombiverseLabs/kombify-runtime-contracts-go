# Compatibility policy

Compatibility is proven in the repository that owns each side of the change.
The contract module proves stable wire behavior and generation lineage. Every
active consumer proves the exact immutable module checksum and the fixtures it
actually decodes, validates, or emits. A green compile alone is insufficient.

## Consumer evidence

A consumer compatibility gate must:

1. parse its lock using the closed `consumer-lock.schema.json` shape;
2. require the exact module version in `go.mod` and the matching module-zip
   `h1:` checksum in `go.sum`;
3. require a sorted, duplicate-free list of only the imported public packages;
4. hash every fixture from a repository-relative plain-file path;
5. strictly decode each fixture and exercise the consumer's relevant contract
   validation or deterministic digest behavior; and
6. be selected by the consumer's affected test planner when the lock, fixture,
   `go.mod`, `go.sum`, or contract-facing implementation changes.

The module checksum is the exact source-content identity. A branch, mutable
tag, another checkout, or a compile against an unrecorded replacement is not
compatibility evidence.

## Additive change evidence

A change is additive only when all of the following remain true:

- existing exported identifiers, JSON names, enum values, API versions,
  validation outcomes, digest inputs, and Golden bytes are unchanged;
- any new JSON field is optional for both writer and reader, and unchanged
  consumer fixtures still pass;
- generated contracts retain their owning schema and output lineage; and
- every active consumer passes its unchanged lock and fixtures against the
  proposed immutable module version.

The producer adds or updates focused Golden/fuzz seeds for the new behavior.
Each consumer updates its version/checksum lock only after its own fixture gate
passes. The evidence is a minor prerelease change during `v0.x`.

## Coordinated breaking-change evidence

Removal or rename of an exported identifier or JSON field, a changed enum/API
version/validation outcome/digest input, a newly required field, or changed
existing Golden bytes is coordinated and breaking. It requires:

- a new contract version with before/after Golden and fuzz evidence;
- one direct migration in every active consumer, with its lock and fixtures
  updated in the same consumer change;
- exact producer and consumer SHAs recorded in the release handoff; and
- proof that no retired alias, facade, or duplicate package was introduced.

During `v0.x` this is at least a minor-version cutover. It must never be hidden
behind a compatibility wrapper or a broad test suite that runs after merge.
