# Development

Install the pinned tools and run the complete gate:

```bash
mise install
mise run check
```

For fast iteration, run only the affected contract package:

```bash
mise run test:affected -- runtimelease
mise run test:affected -- providerexecutor/v1beta1
```

Keep every command below ten minutes. If a contract test becomes slow, split it
into deterministic package-level phases instead of extending its timeout.

`runtimeinventory` changes start in the Techstack inventory schema.
`stackaction` changes start in StackKits CUE and are regenerated there. Commit
the verified bundle and source/output hashes together with the consumer
projection.

Contract-facing changes also run the root compatibility corpus. A consumer
must copy only relevant secret-free seeds into its own repository, lock their
SHA-256 values together with the exact module `h1:` checksum, and select that
focused gate for dependency or contract-facing changes. Do not satisfy a
consumer upgrade with a compile-only check.
