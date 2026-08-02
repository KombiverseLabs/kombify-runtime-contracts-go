# Agent Instructions

<!-- BEGIN GENERATED: planning-policy kombify-agent-policy-sync -->
> Generated from `AGENTS.md` in the kombify workspace root. Do not edit this
> block in child repos; update the root policy and run
> `mise run agents:planning:sync`.

## Planning System Policy

- The workspace root `AGENTS.md` `## Planning System Policy` section is the
  canonical source for generated planning-policy blocks in repo-local
  `AGENTS.md` files. Edit the workspace root policy first, then run
  `mise run agents:planning:sync`. Do not hand-edit the generated blocks in
  child repos.
- Linear is the canonical portfolio and roadmap planning system: high-level epics, cross-repo priorities, phase gates, ownership, and blockers. It is the single source of truth for what to build and when. Full taxonomy and workflow: `LINEAR-PLANNING-STANDARD.md`.
- Workspace: Kombiverse Labs (team KOM). Every Development-project issue carries exactly one `area:*` label; detailed AI component tracking lives in the separate kombify-AI project.
- Repo-local `ROADMAP.md` and optional `docs/roadmap/v0.x.0-*.md` files remain the canonical repo milestone scope and release-gate documents.
- Beads is the canonical execution tracker inside each repo. Keep detailed tasks, subtasks, bugs, bugfixes, dependencies, and technical-depth follow-ups in Beads only.
- Linear and Beads are cross-referenced, not synced: a Linear issue may cite Beads IDs and a Beads issue may cite a Linear ID. Either can exist without the other.
- Check/update Linear at session boundaries (start and end), not on every Beads operation. Do not recreate one-way or bidirectional roadmap syncs between Beads, Linear, and repo docs beyond the two sanctioned generated read views below.
- Sanctioned one-way read views (User-Decision 2026-06-10, see `STANDARDS_ENFORCEMENT.md`): (1) `roadmap-sync` mirrors each ROADMAP.md milestone into one Linear issue (`[<repo>] M<r> · v0.x.0 — <Name>`, label `roadmap:milestone`; the derived rank M1..M5 is the execution order of the active milestones); (2) `roadmap-open-issues` renders open Beads issues into the marked `## Open Issues` block inside ROADMAP.md. Both are derived views — never edit them manually, never sync back.
- Session close with milestone-relevant work: update the Scope/Exit-gate checkboxes in the touched repo's repo-local `ROADMAP.md`, then run `mise run roadmap:update -- -Repo <repo>` from the workspace root (refreshes the Open-Issues block; add `-Sync` to push the Linear mirror).
<!-- END GENERATED: planning-policy kombify-agent-policy-sync -->

This repository is the public, dependency-light Go contract module for the
Kombify runtime boundary. Read the canonical workspace
`PLATFORM-STRATEGY.md`, `PLATFORM-ARCHITECTURE-TARGET.md`,
`PUBLIC-BETA-PLAN.md`, and repository manifest before changing a contract.

## Boundaries

- Keep exactly four public contract surfaces: `runtimelease`,
  `providerexecutor/v1beta1`, `runtimeinventory`, and `stackaction`.
- Do not add provider SDKs, credentials, connection-profile implementations,
  adapter registries, databases, durable product state, billing/product policy,
  or hosted-service clients.
- Do not add aliases or wrappers for `vmlease`, `runtimeaction`,
  `serverruntime`, or a previous provider-executor wire.
- `runtimeinventory` is generated from the Techstack inventory schema.
- `stackaction` is generated from the canonical StackKits CUE contract. Edit
  the CUE source and generator in StackKits, then copy the verified generation
  bundle; never hand-edit the generated Go projection.
- Public source, tests, docs, and fixtures must contain no private dependency,
  URL, secret, PII, or production data.

## Commands

```bash
mise run check
mise run test:affected -- runtimelease
mise run test:affected -- providerexecutor/v1beta1
mise run test:affected -- runtimeinventory
mise run test:affected -- stackaction
```

Use Beads for repo-local execution work, ROADMAP.md for milestone scope, and
Linear for cross-repository priority or blockers. Completed work must be
committed and pushed from an isolated branch.
