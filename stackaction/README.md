# stackaction

`stackaction` is generated from the canonical StackKits CUE action contract.
The checked-in generation bundle contains:

- `stackaction_gen.go`: public Go wire projection;
- `contract.ir.json`: deterministic intermediate representation;
- `openapi.yaml`: matching OpenAPI component projection;
- `manifest.json`: generator version, wire version, CUE source SHA-256, and
  SHA-256 for every output.

The package is not a second schema authority. Update the CUE source and
generator in StackKits, run its `contracts:stackaction:check` task, and copy the
verified bundle byte-for-byte. `generation_test.go` rejects local drift.
