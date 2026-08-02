# Kombify Runtime Contracts for Go

[![Latest release: v0.1.4](https://img.shields.io/badge/release-v0.1.4-blue.svg)](VERSION)
[![License: Apache-2.0](https://img.shields.io/badge/license-Apache--2.0-blue.svg)](LICENSE)

Dependency-light, public Go contracts for the Kombify runtime boundary. The
module publishes stable type identities and deterministic JSON fixtures; it is
not a service, control plane, provider adapter, or product-state store.

## Packages

| Import path | Contract authority |
| --- | --- |
| `runtimelease` | Secret-free lease identity, validation, enrollment, and signed snapshot DTOs |
| `providerexecutor/v1beta1` | Provider-neutral commands, receipts, evidence, digesting, and lifecycle validation |
| `runtimeinventory` | Secret-free runtime inventory read model generated from the Techstack inventory schema |
| `stackaction` | StackKits action wire generated from the canonical StackKits CUE contract |

Techstack owns lease state, runtime inventory behavior, provider adapter
selection and execution, durable resource custody, and product policy.
StackKits owns the CUE action schema and lifecycle behavior. This module owns
only the published Go contract projections and compatibility fixtures.

The module intentionally contains no provider SDK, credential handling,
adapter registry, database, durable product state, billing policy, or legacy
`vmlease`, `runtimeaction`, and `serverruntime` compatibility packages.

## Install

```bash
go get github.com/KombiverseLabs/kombify-runtime-contracts-go@v0.1.4
```

`v0.1.4` is the latest immutable public release. Changes on `main` remain
unreleased until a Delivery run binds a new tag to the exact current source.

## Verify

```bash
mise install
mise run check
```

`mise run check` formats-checks, vets, tests, builds, verifies Golden JSON, and
enforces the public dependency and ownership boundary.

## Compatibility

The module follows Semantic Versioning. While the module is below `v1.0.0`,
wire-compatible additions use a minor release and breaking wire or Go API
changes require a coordinated minor-version cutover. Published JSON field
names, enum values, digest rules, and Golden fixtures are compatibility
contracts. No retired package path is retained as an alias or wrapper.

Current implementation evidence is in [STATUS.md](STATUS.md), future release
scope is in [ROADMAP.md](ROADMAP.md), and package architecture and API rules
are in [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) and
[docs/API.md](docs/API.md).

## License

Apache-2.0. See [LICENSE](LICENSE).
