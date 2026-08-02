# providerexecutor

This package implements the neutral `providerexecutor/v1beta1` contract for
infrastructure-provider lifecycle execution. It is the only provider-control
wire and is not a service contract.

## Ownership

- **Techstack** authorizes commands, selects/configures adapters, owns leases,
  persists the append-only lifecycle and resource ledger, and retains cleanup
  handles until provider-native absence is proven.
- **providerexecutor/v1beta1** defines the transport-independent `Executor`
  interface, command/receipt wire types, deterministic idempotency hashes,
  resource graph, evidence contract, and lifecycle transition rules. The
  package is the public Go contract authority.
- A bounded conformance harness may implement or consume the same contract for simulations and
  tests. It is not a required network hop or lifecycle authority.
- **StackKits** does not import provider adapters or provider compatibility. It
  receives an external host binding and reports OS/host conformance separately.

## Safety invariants

- Commands contain hierarchical server-side lookup references whose identifiers
  remain opaque to the shared module, never raw credentials: `custody://`,
  `provider-connection://`, and
  `desired-spec://`. Every command binds the exact positive `lease_revision`,
  durable `runtime_server_id`, canonical lowercase `resource_generation_id`
  UUID, stable vendor `provider_id`, TechStack `adapter_id`, capability
  snapshot, execution profile, custody, connection, resource graph, and
  positive ledger revision. Receipts and evidence repeat the relevant fences;
  command digests, replay validation, evidence subjects, and receipt identity
  reject drift. `SealCommand` never invents a generation. Operation identity
  includes the generation, so a later generation cannot address an older
  ledger.
- Reusing an idempotency key within the same tenant, lease, generation, and
  operation scope with different command content fails closed.
- Target and receipt resource graphs are closed, parent-complete, acyclic, and
  bind ownership plus the required delete disposition. One envelope is bounded
  to 128 targets, 256 resources, 32 evidence entries per resource, and 512
  evidence entries total. Raw cardinality and byte bounds are checked before
  normalization, copying, sorting, or marshaling. Revisions and receipt
  sequences are positive and capped at JSON's cross-runtime safe integer
  maximum. Observation and cleanup values use a closed
  compatibility matrix. `SealCommand` derives `resource_graph_hash` from the
  normalized targets; callers cannot authorize an unrelated placeholder hash.
- Plan receipts are always resource-free. Failed mutating operations may retain
  partial resource graphs; known native references remain available for cleanup
  rather than escaping on an error return.
- Denied receipts are resource-free and may be appended only from the durable
  `requested` head through `DenyReceipt`, before an executor invocation.
  Executors cannot return `denied`. Once `accepted` records invocation custody,
  even an ambiguous result with no known handle must be `failed` and reconciled;
  denial is never historical side-effect evidence.
- Successful observe and decommission receipts exactly cover their command
  targets. Successful reconcile receipts retain exactly the target graph.
- Absence requires unique, definitive, attested provider API/event evidence
  bound to the operation, lease revision, runtime server, provider, capability
  snapshot, binding, native reference, connection, execution profile, graph,
  and subject. Collection time must fall between command request and receipt
  issue, and `EvidenceVerifier` must accept the attestation. Verifiers receive
  the caller context, must honor cancellation, and typed-nil implementations
  fail closed. Adapter-local assertions alone do not complete cleanup.
- TechStack's coordinator creates the sole initial receipt: sequence 1 is
  exactly `pending/requested`, has no predecessor or resources, and is persisted
  before invoking an Executor. Each executor call receives an
  `ExecutionRequest` containing the immutable command and the validated prior
  receipt, so it can return exactly one next transition with the retained graph
  and evidence history. Executors return unchained `ExecutionResult` values,
  not receipts; the coordinator supplies its ledger head to `AssembleReceipt`,
  then persists the sealed candidate. Receipts form a hash chain through
  `Sequence` and `PreviousReceiptDigest`. Every receipt records
  `phase_entered_at`; transitions set it to the receipt issue time and
  same-phase polling preserves it exactly. `ValidateAppend` enforces exact
  sequence increments, monotone issue time, legal phase transitions, and
  resource/evidence continuity. Terminal decommission absence evidence must
  have been collected at or after the preceding `absence_pending` entry began.
  Pending `accepted`, `resources_bound`,
  `delete_accepted`, and `absence_pending` phases may append safe same-phase
  polling receipts; requested and terminal phases may not self-loop.
- `Executor.Execute` returns one `ExecutionResult`; provider failures use a
  failed result with a stable secret-free `{code,retryable}` reason and any
  partial resource bindings. Human diagnostics and provider responses stay in
  TechStack-owned logs/evidence and never enter the receipt. The
  TechStack coordinator alone assembles the next receipt from that result and
  its durable chain head. The `provider.*` namespace is closed to auth,
  permission, invalid spec, quota, rate limit, conflict, not found, transient,
  timeout, partial create, and cleanup required; adapter-internal diagnostics
  cannot create additional provider wire categories.
- The shared package includes no provider SDK, database, adapter registry, or
  vendor-specific configuration.

Evidence is append-only per resource. An older evidence envelope remains bound
to its own observation (for example `present`) after the current resource
observation becomes `unknown` or `absent`; validation never rewrites history.
The newest definitive evidence must determine the current observation; equal
collection times with conflicting observations are rejected. Current absence
still requires a new definitive provider-native absence proof,
and successful observe requires definitive evidence for the current observation.

`PhaseAbsent` is operation-specific. A successful observe receipt uses it to
state that every exact target is definitively absent; it does not claim that a
delete was requested or accepted. A successful decommission reaches the same
terminal phase only through `delete_accepted -> absence_pending -> absent`,
with exact target coverage, fresh post-`absence_pending` provider evidence,
and complete cleanup.

## Receipt validation

Call `SealCommand` before execution and persist/execute only its canonical
result. The TechStack coordinator creates and persists the chain head with
`InitialReceipt`, validates the `ExecutionRequest` before invoking an adapter,
and passes the returned `ExecutionResult` to `AssembleReceipt`. That helper
seals the next entry and runs `ValidateAppend` against the supplied durable
head; only the returned receipt may be persisted. Admission failures use
`DenyReceipt` while the head is still `requested`. There is intentionally no
public standalone receipt-sealing helper: a post-initial receipt without its
prior head cannot be authorized for append. `Receipt.ValidateFor` is the
standalone read/replay integrity check, not authority to append an entry with
an unverified predecessor. Direct validation rejects any value whose
whitespace, digest case, ordering, or timestamps are not already canonical.
Receipt validation, assembly, and append calls take a context. The
TechStack-owned verifier may be nil only when the receipt and retained history
contain no absence claim.
