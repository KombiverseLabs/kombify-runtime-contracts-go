# Changelog

All notable changes to this module are documented here.

## Unreleased

- Bind the `runtimeinventory` projection to its canonical Techstack schema with
  an exact source/output digest manifest instead of a parallel hand-maintained
  field map.
- Bind release dispatches to the exact centrally pinned Delivery compiler and
  correlate the authoritative adapter run by release identity plus plan digest.

## 0.1.4 - 2026-08-02

- Correct the source version to `0.1.4` after the immutable historical `v0.1.3`
  tag was found to contain stale `0.1.0` source metadata.
- Fail closed when a Delivery release version or tag differs from the exact
  `.kombify/VERSION` committed at the authorized source SHA.
