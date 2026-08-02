# API Contract

The public API is the exported Go surface of the four packages and their
checked-in JSON fixtures.

## Stability rules

- JSON field names, enum values, API/wire version strings, digest algorithms,
  and Golden fixture bytes are compatibility-sensitive.
- Closed validators reject unknown or malformed authority-bearing values.
- Additive Go fields require explicit wire-compatibility review; removals,
  renames, enum changes, and digest changes require a coordinated SemVer
  cutover.
- `providerexecutor/v1beta1` is the only provider control version. No root
  facade or previous wire is exported.
- `runtimeinventory` and `stackaction` generated files are never edited by
  hand; changes start in their owning schemas.

Package documentation contains the detailed invariants next to the exported
types and validators.
