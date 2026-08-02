# Changelog

All notable changes to this module are documented here.

## Unreleased

- Correct the source version to `0.1.4` after the immutable historical `v0.1.3`
  tag was found to contain stale `0.1.0` source metadata.
- Fail closed when a Delivery release version or tag differs from the exact
  `.kombify/VERSION` committed at the authorized source SHA.
