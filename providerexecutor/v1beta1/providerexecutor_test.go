package providerexecutor

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"
)

var contractNow = time.Date(2026, 7, 17, 14, 0, 0, 0, time.UTC)

const contractResourceGenerationID = "550e8400-e29b-41d4-a716-446655440000"

func TestSealCommandAndValidateReplay(t *testing.T) {
	command := mustCommand(t, validCommand(OperationProvision))
	if command.APIVersion != APIVersion || command.OperationID == "" || command.CommandDigest == "" {
		t.Fatalf("sealed command missing identity: %+v", command)
	}
	if len(command.OperationID) != len("op_")+32 {
		t.Fatalf("operation identity length = %d, want 128-bit hexadecimal identity", len(command.OperationID))
	}
	if err := command.Validate(); err != nil {
		t.Fatalf("Validate: %v", err)
	}
	if err := ValidateReplay(command, command); err != nil {
		t.Fatalf("ValidateReplay exact replay: %v", err)
	}

	replayInput := validCommand(OperationProvision)
	replayInput.AdapterID = "different-adapter"
	replay := mustCommand(t, replayInput)
	if err := ValidateReplay(command, replay); !errors.Is(err, ErrIdempotencyConflict) {
		t.Fatalf("ValidateReplay error = %v, want ErrIdempotencyConflict", err)
	}

	generationReplayInput := validCommand(OperationProvision)
	generationReplayInput.ResourceGenerationID = "f47ac10b-58cc-4372-a567-0e02b2c3d479"
	generationReplay := mustCommand(t, generationReplayInput)
	if generationReplay.OperationID == command.OperationID {
		t.Fatal("operation identity did not change across resource generations")
	}
	if err := ValidateReplay(command, generationReplay); !errors.Is(err, ErrIdempotencyConflict) {
		t.Fatalf("generation replay error = %v, want ErrIdempotencyConflict", err)
	}
}

func TestResourceGenerationPropagatesThroughReceiptChain(t *testing.T) {
	command := mustCommand(t, validCommand(OperationPlan))
	initial, err := InitialReceipt(command, contractNow)
	if err != nil {
		t.Fatalf("InitialReceipt: %v", err)
	}
	if initial.ResourceGenerationID != command.ResourceGenerationID {
		t.Fatalf("initial resource_generation_id = %q, want %q", initial.ResourceGenerationID, command.ResourceGenerationID)
	}
	next, err := AssembleReceipt(
		context.Background(),
		ExecutionRequest{Command: command, Previous: initial},
		ExecutionResult{Status: StatusPending, Phase: PhaseAccepted},
		contractNow.Add(time.Second),
		nil,
	)
	if err != nil {
		t.Fatalf("AssembleReceipt: %v", err)
	}
	if next.ResourceGenerationID != command.ResourceGenerationID {
		t.Fatalf("assembled resource_generation_id = %q, want %q", next.ResourceGenerationID, command.ResourceGenerationID)
	}

	drifted := next
	drifted.ResourceGenerationID = "f47ac10b-58cc-4372-a567-0e02b2c3d479"
	drifted = resealReceipt(t, drifted)
	if err := drifted.ValidateFor(context.Background(), command, nil); !errors.Is(err, ErrInvalidReceipt) {
		t.Fatalf("generation-drifted receipt error = %v, want ErrInvalidReceipt", err)
	}
}

func TestProviderExecutorV1Beta1GoldenJSON(t *testing.T) {
	command := mustCommand(t, validCommand(OperationPlan))
	receipt, err := InitialReceipt(command, contractNow)
	if err != nil {
		t.Fatalf("InitialReceipt: %v", err)
	}
	payload, err := json.MarshalIndent(struct {
		Command Command `json:"command"`
		Receipt Receipt `json:"receipt"`
	}{Command: command, Receipt: receipt}, "", "  ")
	if err != nil {
		t.Fatalf("MarshalIndent: %v", err)
	}
	const want = `{
  "command": {
    "api_version": "providerexecutor/v1beta1",
    "operation_id": "op_72f673447b51fa7a84e662ea1922edf4",
    "operation": "plan",
    "tenant_id": "tenant-1",
    "lease_id": "lease-1",
    "lease_revision": 11,
    "runtime_server_id": "server-1",
    "resource_generation_id": "550e8400-e29b-41d4-a716-446655440000",
    "idempotency_key": "request-1",
    "provider_id": "centron",
    "adapter_id": "managed-compute",
    "custody_ref": "custody://tenant-1/provider-access",
    "custody_hash": "sha256:21a7f61c15cef4ddedae076d6b7393f1d3d4a9b5c870df60ca0c59b195ef1602",
    "connection_ref": "provider-connection://tenant-1/managed-compute",
    "connection_hash": "sha256:b38d9d168c3aedf156f4f249b81adaef4b738790510573f57b502cca0c35f16f",
    "capability_snapshot_hash": "sha256:3fc0e5c4c484797a0035141c53252fba96ec5923ec719b5c73d2b9078b31bf08",
    "execution_profile_hash": "sha256:1900eab6c028483d7126599ee6f50de0d27907b5c65fa90524580b4b0f9852b0",
    "resource_graph_hash": "sha256:4f53cda18c2baa0c0354bb5f9a3ecbe5ed12ab4d8e11ba873c2f11161202b945",
    "ledger_revision": 7,
    "desired_spec_ref": "desired-spec://techstack/leases/lease-1",
    "desired_spec_hash": "sha256:d4f02eaafd1a9e9de7d10972ca8e47fa7a985825c3c9c1e249c72683cb3e4f19",
    "requested_at": "2026-07-17T14:00:00Z",
    "command_digest": "sha256:701f41e84ee1c8e220ad388eddf81294be9dc55b4bef01d6c62f6023ba837937"
  },
  "receipt": {
    "api_version": "providerexecutor/v1beta1",
    "operation_id": "op_72f673447b51fa7a84e662ea1922edf4",
    "operation": "plan",
    "tenant_id": "tenant-1",
    "lease_id": "lease-1",
    "lease_revision": 11,
    "runtime_server_id": "server-1",
    "resource_generation_id": "550e8400-e29b-41d4-a716-446655440000",
    "idempotency_key": "request-1",
    "provider_id": "centron",
    "adapter_id": "managed-compute",
    "capability_snapshot_hash": "sha256:3fc0e5c4c484797a0035141c53252fba96ec5923ec719b5c73d2b9078b31bf08",
    "command_digest": "sha256:701f41e84ee1c8e220ad388eddf81294be9dc55b4bef01d6c62f6023ba837937",
    "desired_spec_hash": "sha256:d4f02eaafd1a9e9de7d10972ca8e47fa7a985825c3c9c1e249c72683cb3e4f19",
    "sequence": 1,
    "status": "pending",
    "phase": "requested",
    "phase_entered_at": "2026-07-17T14:00:00Z",
    "issued_at": "2026-07-17T14:00:00Z",
    "receipt_digest": "sha256:82987078e7add3a0808922507c63528dbae706b262cf2ddcc29c44767d555cf9"
  }
}`
	if string(payload) != want {
		t.Fatalf("providerexecutor/v1beta1 JSON drift:\n%s", payload)
	}
}

func TestExecutionFencesPropagateAndRejectDrift(t *testing.T) {
	command := mustCommand(t, validCommand(OperationObserve))
	receipt, err := InitialReceipt(command, contractNow)
	if err != nil {
		t.Fatalf("InitialReceipt: %v", err)
	}
	if receipt.LeaseRevision != command.LeaseRevision || receipt.RuntimeServerID != command.RuntimeServerID ||
		receipt.ProviderID != command.ProviderID || receipt.CapabilitySnapshotHash != command.CapabilitySnapshotHash {
		t.Fatalf("receipt execution fences did not propagate: %+v", receipt)
	}

	for _, tt := range []struct {
		name   string
		mutate func(*Receipt)
	}{
		{"lease revision", func(r *Receipt) { r.LeaseRevision++ }},
		{"runtime server", func(r *Receipt) { r.RuntimeServerID = "server-2" }},
		{"provider", func(r *Receipt) { r.ProviderID = "ionos" }},
		{"capability snapshot", func(r *Receipt) { r.CapabilitySnapshotHash = testDigest("other-capabilities") }},
	} {
		t.Run("receipt "+tt.name, func(t *testing.T) {
			candidate := receipt
			tt.mutate(&candidate)
			candidate = resealReceipt(t, candidate)
			if err := candidate.ValidateFor(context.Background(), command, nil); !errors.Is(err, ErrInvalidReceipt) {
				t.Fatalf("ValidateFor error = %v, want ErrInvalidReceipt", err)
			}
		})
	}

	for _, tt := range []struct {
		name   string
		mutate func(*Command)
	}{
		{"lease revision", func(c *Command) { c.LeaseRevision++ }},
		{"runtime server", func(c *Command) { c.RuntimeServerID = "server-2" }},
		{"provider", func(c *Command) { c.ProviderID = "ionos" }},
		{"capability snapshot", func(c *Command) { c.CapabilitySnapshotHash = testDigest("other-capabilities") }},
	} {
		t.Run("replay "+tt.name, func(t *testing.T) {
			input := validCommand(OperationObserve)
			tt.mutate(&input)
			replay := mustCommand(t, input)
			if err := ValidateReplay(command, replay); !errors.Is(err, ErrIdempotencyConflict) {
				t.Fatalf("ValidateReplay error = %v, want ErrIdempotencyConflict", err)
			}
		})
	}
}

func TestEvidenceFenceGoldenJSON(t *testing.T) {
	command := mustCommand(t, validCommand(OperationObserve))
	evidence := observationEvidence(command, command.Targets[0], ObservationPresent, "golden")
	payload, err := json.MarshalIndent(evidence, "", "  ")
	if err != nil {
		t.Fatalf("MarshalIndent: %v", err)
	}
	const want = `{
  "ref": "provider-evidence://operations/golden",
  "digest": "sha256:b3955036ec3b0db6dd68cf705c3ac47cd52ded7b2db54c533edb7c2f22c2e064",
  "source": "provider_api",
  "operation_id": "op_cf245933106d2d3e299ffefbf3e4d0b5",
  "lease_revision": 11,
  "runtime_server_id": "server-1",
  "provider_id": "centron",
  "capability_snapshot_hash": "sha256:3fc0e5c4c484797a0035141c53252fba96ec5923ec719b5c73d2b9078b31bf08",
  "binding_id": "server",
  "native_ref_hash": "sha256:abb8849f19b1cc519719a7362e35106e0607e7da6d1914ce1c351e02ee48e327",
  "connection_hash": "sha256:b38d9d168c3aedf156f4f249b81adaef4b738790510573f57b502cca0c35f16f",
  "execution_profile_hash": "sha256:1900eab6c028483d7126599ee6f50de0d27907b5c65fa90524580b4b0f9852b0",
  "resource_graph_hash": "sha256:ae416750a66ecac30a0298ef5e5b06c117e12e29f324e6cfb74bbd80668eb560",
  "subject_hash": "sha256:c1b16b83b2e87af6d00b32ec154907bf13854d23800aa4af4c6b7bd33e7beb17",
  "observation": "present",
  "definitive": true,
  "attestation_ref": "provider-attestation://operations/golden",
  "attestation_digest": "sha256:64ba8cbe902076fbb59453c6000a921d43234e4a9d3eca3fb941fc626465151d",
  "collected_at": "2026-07-17T14:00:30Z"
}`
	if string(payload) != want {
		t.Fatalf("evidence JSON drift:\n%s", payload)
	}
}

func TestEvidenceExecutionFencesRejectDrift(t *testing.T) {
	command := mustCommand(t, validCommand(OperationObserve))
	target := command.Targets[0]
	resource := resourceBinding(target, ObservationPresent, CleanupRequired)
	resource.Evidence = []Evidence{observationEvidence(command, target, ObservationPresent, "fenced")}
	base := Receipt{Status: StatusSucceeded, Phase: PhasePresent, Resources: []ResourceBinding{resource}}
	if _, err := sealReceipt(t, command, base, nil); err != nil {
		t.Fatalf("valid fenced evidence: %v", err)
	}
	for _, tt := range []struct {
		name   string
		mutate func(*Evidence)
	}{
		{"lease revision", func(e *Evidence) { e.LeaseRevision++ }},
		{"runtime server", func(e *Evidence) { e.RuntimeServerID = "server-2" }},
		{"provider", func(e *Evidence) { e.ProviderID = "ionos" }},
		{"capability snapshot", func(e *Evidence) { e.CapabilitySnapshotHash = testDigest("other-capabilities") }},
	} {
		t.Run(tt.name, func(t *testing.T) {
			candidate := base
			candidate.Resources = cloneResources(base.Resources)
			tt.mutate(&candidate.Resources[0].Evidence[0])
			if _, err := sealReceipt(t, command, candidate, nil); !errors.Is(err, ErrInvalidReceipt) {
				t.Fatalf("evidence drift error = %v, want ErrInvalidReceipt", err)
			}
		})
	}
}

func TestValidateRejectsNonCanonicalCommand(t *testing.T) {
	input := validCommand(OperationObserve)
	input.Targets = []ResourceTarget{resourceTarget("server"), resourceTarget("network")}
	sealed := mustCommand(t, input)
	tests := []struct {
		name   string
		mutate func(*Command)
	}{
		{"whitespace", func(c *Command) { c.TenantID = " " + c.TenantID }},
		{"resource generation case", func(c *Command) { c.ResourceGenerationID = strings.ToUpper(c.ResourceGenerationID) }},
		{"digest case", func(c *Command) { c.ConnectionHash = strings.ToUpper(c.ConnectionHash) }},
		{"target order", func(c *Command) { c.Targets[0], c.Targets[1] = c.Targets[1], c.Targets[0] }},
		{"target whitespace", func(c *Command) { c.Targets[0].NativeRef += " " }},
		{"timestamp location", func(c *Command) { c.RequestedAt = c.RequestedAt.In(time.FixedZone("review", 3600)) }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			candidate := sealed
			candidate.Targets = append([]ResourceTarget(nil), sealed.Targets...)
			tt.mutate(&candidate)
			if isCanonicalCommand(candidate) {
				t.Fatal("mutated command unexpectedly canonical")
			}
			if err := candidate.Validate(); !errors.Is(err, ErrInvalidCommand) {
				t.Fatalf("Validate error = %v, want ErrInvalidCommand", err)
			}
		})
	}
}

func TestSealCommandReturnsCanonicalEnvelope(t *testing.T) {
	command := validCommand(OperationObserve)
	command.APIVersion = " " + APIVersion + " "
	command.Operation = Operation(" observe ")
	command.TenantID = " tenant-1 "
	command.ResourceGenerationID = strings.ToUpper(command.ResourceGenerationID)
	command.CustodyHash = strings.ToUpper(command.CustodyHash)
	command.ConnectionRef = " " + command.ConnectionRef + " "
	command.RequestedAt = command.RequestedAt.In(time.FixedZone("review", 3600))
	command.Targets = []ResourceTarget{resourceTarget("server"), resourceTarget("network")}
	command.Targets[0].NativeRef += " "
	command.Targets[1].OwnershipHash = strings.ToUpper(command.Targets[1].OwnershipHash)
	sealed := mustCommand(t, command)
	if !isCanonicalCommand(sealed) {
		t.Fatalf("SealCommand returned non-canonical command: %+v", sealed)
	}
	if sealed.Targets[0].BindingID != "network" || sealed.RequestedAt.Location() != time.UTC {
		t.Fatalf("SealCommand did not sort targets and canonicalize time: %+v", sealed)
	}
	if sealed.ResourceGenerationID != contractResourceGenerationID {
		t.Fatalf("resource_generation_id = %q, want canonical %q", sealed.ResourceGenerationID, contractResourceGenerationID)
	}
}

func TestResourceGraphHashIsDerivedAndCannotDrift(t *testing.T) {
	command := validCommand(OperationObserve)
	command.Targets = []ResourceTarget{resourceTarget("server"), resourceTarget("network")}
	command.ResourceGraphHash = testDigest("caller-placeholder")
	sealed := mustCommand(t, command)
	expected, err := ComputeResourceGraphHash(sealed.Targets)
	if err != nil {
		t.Fatalf("ComputeResourceGraphHash: %v", err)
	}
	if sealed.ResourceGraphHash != expected || sealed.ResourceGraphHash == testDigest("caller-placeholder") {
		t.Fatalf("resource graph hash = %q, want derived %q", sealed.ResourceGraphHash, expected)
	}

	drifted := sealed
	drifted.ResourceGraphHash = testDigest("different-graph")
	drifted.CommandDigest, err = ComputeCommandDigest(drifted)
	if err != nil {
		t.Fatalf("ComputeCommandDigest: %v", err)
	}
	if err := drifted.Validate(); !errors.Is(err, ErrInvalidCommand) {
		t.Fatalf("drifted graph Validate error = %v, want ErrInvalidCommand", err)
	}

	reordered := []ResourceTarget{sealed.Targets[1], sealed.Targets[0]}
	reorderedHash, err := ComputeResourceGraphHash(reordered)
	if err != nil {
		t.Fatalf("ComputeResourceGraphHash(reordered): %v", err)
	}
	if reorderedHash != expected {
		t.Fatalf("reordered graph hash = %q, want %q", reorderedHash, expected)
	}
}

func TestCommandRequiresAuthorityHashesRevisionAndFieldSpecificRefs(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*Command)
	}{
		{"custody ref", func(c *Command) { c.CustodyRef = "" }},
		{"lease revision", func(c *Command) { c.LeaseRevision = 0 }},
		{"runtime server", func(c *Command) { c.RuntimeServerID = "" }},
		{"resource generation", func(c *Command) { c.ResourceGenerationID = "" }},
		{"nil resource generation", func(c *Command) { c.ResourceGenerationID = "00000000-0000-0000-0000-000000000000" }},
		{"custody scheme", func(c *Command) { c.CustodyRef = c.ConnectionRef }},
		{"custody hash", func(c *Command) { c.CustodyHash = "" }},
		{"connection ref", func(c *Command) { c.ConnectionRef = "custody://tenant-1/provider" }},
		{"connection hash", func(c *Command) { c.ConnectionHash = "" }},
		{"provider", func(c *Command) { c.ProviderID = "" }},
		{"capability snapshot hash", func(c *Command) { c.CapabilitySnapshotHash = "" }},
		{"profile hash", func(c *Command) { c.ExecutionProfileHash = "" }},
		{"ledger revision", func(c *Command) { c.LedgerRevision = 0 }},
		{"desired spec scheme", func(c *Command) { c.DesiredSpecRef = "provider-connection://tenant-1/spec" }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			command := validCommand(OperationProvision)
			tt.mutate(&command)
			if _, err := SealCommand(command); !errors.Is(err, ErrInvalidCommand) {
				t.Fatalf("SealCommand error = %v, want ErrInvalidCommand", err)
			}
		})
	}
}

func TestServerLookupReferencesUseStrictHierarchicalGrammar(t *testing.T) {
	valid := []struct {
		name   string
		value  string
		scheme string
	}{
		{"custody", "custody://tenant-1/provider-access", custodyRefScheme},
		{"connection", "provider-connection://tenant-1/managed-compute", connectionRefScheme},
		{"desired spec", "desired-spec://techstack/leases/lease-1", desiredSpecRefScheme},
		{"evidence", "provider-evidence://operations/id", evidenceRefScheme},
		{"attestation", "provider-attestation://operations/id", attestationRefScheme},
	}
	for _, tt := range valid {
		t.Run("accepts "+tt.name, func(t *testing.T) {
			if err := validateLookupRef(tt.name, tt.value, tt.scheme); err != nil {
				t.Fatalf("validateLookupRef(%q): %v", tt.value, err)
			}
		})
	}

	invalid := []string{
		"custody:raw-secret",
		"custody://raw-secret",
		"custody://user@tenant-1/provider-access",
		"custody://tenant-1:443/provider-access",
		"custody://tenant-1/provider%2Faccess",
		"custody://tenant-1/provider-access?token=secret",
		"custody://tenant-1/provider-access#fragment",
		"custody://tenant-1/provider-access#",
		"custody:///provider-access",
		"custody://tenant-1/",
		"custody://tenant-1//provider-access",
		"custody://tenant-1/./provider-access",
		"custody://tenant-1/../provider-access",
		"custody://tenant-1./provider-access",
		"custody://ténant-1/provider-access",
		"custody://tenant-1/-provider-access",
	}
	for _, value := range invalid {
		t.Run("rejects "+value, func(t *testing.T) {
			if err := validateLookupRef("custody_ref", value, custodyRefScheme); !errors.Is(err, ErrInvalidCommand) {
				t.Fatalf("validateLookupRef(%q) error = %v, want ErrInvalidCommand", value, err)
			}
		})
	}
}

func TestCommandRejectsInvalidUTF8InEveryStringEnvelope(t *testing.T) {
	bad := string([]byte{0xff})
	tests := []struct {
		name   string
		mutate func(*Command)
	}{
		{"api version", func(c *Command) { c.APIVersion = bad }},
		{"operation id", func(c *Command) { c.OperationID = bad }},
		{"operation", func(c *Command) { c.Operation = Operation(bad) }},
		{"tenant", func(c *Command) { c.TenantID = bad }},
		{"lease", func(c *Command) { c.LeaseID = bad }},
		{"runtime server", func(c *Command) { c.RuntimeServerID = bad }},
		{"resource generation", func(c *Command) { c.ResourceGenerationID = bad }},
		{"idempotency", func(c *Command) { c.IdempotencyKey = bad }},
		{"provider", func(c *Command) { c.ProviderID = bad }},
		{"adapter", func(c *Command) { c.AdapterID = bad }},
		{"custody ref", func(c *Command) { c.CustodyRef = bad }},
		{"custody hash", func(c *Command) { c.CustodyHash = bad }},
		{"connection ref", func(c *Command) { c.ConnectionRef = bad }},
		{"connection hash", func(c *Command) { c.ConnectionHash = bad }},
		{"capability snapshot hash", func(c *Command) { c.CapabilitySnapshotHash = bad }},
		{"profile hash", func(c *Command) { c.ExecutionProfileHash = bad }},
		{"graph hash", func(c *Command) { c.ResourceGraphHash = bad }},
		{"spec ref", func(c *Command) { c.DesiredSpecRef = bad }},
		{"spec hash", func(c *Command) { c.DesiredSpecHash = bad }},
		{"command digest", func(c *Command) { c.CommandDigest = bad }},
		{"target binding", func(c *Command) { c.Targets[0].BindingID = bad }},
		{"target kind", func(c *Command) { c.Targets[0].Kind = bad }},
		{"target native ref", func(c *Command) { c.Targets[0].NativeRef = bad }},
		{"target parent", func(c *Command) { c.Targets[0].ParentBindingID = bad }},
		{"target ownership", func(c *Command) { c.Targets[0].OwnershipHash = bad }},
		{"target disposition", func(c *Command) { c.Targets[0].Disposition = ResourceDisposition(bad) }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			command := validCommand(OperationReconcile)
			tt.mutate(&command)
			if _, err := SealCommand(command); !errors.Is(err, ErrInvalidCommand) {
				t.Fatalf("SealCommand error = %v, want ErrInvalidCommand", err)
			}
		})
	}
	command := validCommand(OperationProvision)
	command.TenantID = bad
	if _, err := ComputeCommandDigest(command); !errors.Is(err, ErrInvalidCommand) {
		t.Fatalf("ComputeCommandDigest error = %v, want ErrInvalidCommand", err)
	}
}

func TestCommandTargetGraphIsClosedOwnedAndAcyclic(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*Command)
	}{
		{"ownership", func(c *Command) { c.Targets[0].OwnershipHash = "" }},
		{"disposition", func(c *Command) { c.Targets[0].Disposition = "retain" }},
		{"missing parent", func(c *Command) { c.Targets[0].ParentBindingID = "missing" }},
		{"cycle", func(c *Command) {
			c.Targets = append(c.Targets, resourceTarget("network"))
			c.Targets[0].ParentBindingID = "network"
			c.Targets[1].ParentBindingID = "server"
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			command := validCommand(OperationObserve)
			tt.mutate(&command)
			if _, err := SealCommand(command); !errors.Is(err, ErrInvalidCommand) {
				t.Fatalf("SealCommand error = %v, want ErrInvalidCommand", err)
			}
		})
	}
}

func TestCommandRejectsTargetsOnPlanAndSpecOnObserve(t *testing.T) {
	plan := validCommand(OperationPlan)
	plan.Targets = []ResourceTarget{resourceTarget("server")}
	if _, err := SealCommand(plan); !errors.Is(err, ErrInvalidCommand) {
		t.Fatalf("plan targets error = %v, want ErrInvalidCommand", err)
	}
	observe := validCommand(OperationObserve)
	observe.DesiredSpecRef = "desired-spec://techstack/leases/lease-1"
	observe.DesiredSpecHash = testDigest("spec")
	if _, err := SealCommand(observe); !errors.Is(err, ErrInvalidCommand) {
		t.Fatalf("observe spec error = %v, want ErrInvalidCommand", err)
	}
}

func TestSuccessfulProvisionReceiptRetainsResourceGraph(t *testing.T) {
	command := mustCommand(t, validCommand(OperationProvision))
	receipt := mustReceipt(t, command, Receipt{
		Status: StatusSucceeded,
		Phase:  PhasePresent,
		Resources: []ResourceBinding{
			resourceBinding(resourceTarget("server"), ObservationPresent, CleanupRequired),
		},
	}, nil)
	if err := receipt.ValidateFor(context.Background(), command, nil); err != nil {
		t.Fatalf("ValidateFor: %v", err)
	}
}

func TestPlanReceiptsNeverRetainResources(t *testing.T) {
	command := mustCommand(t, validCommand(OperationPlan))
	resource := resourceBinding(resourceTarget("unauthorized-plan-resource"), ObservationPresent, CleanupRequired)
	tests := []struct {
		name    string
		receipt Receipt
	}{
		{
			name: "pending",
			receipt: Receipt{
				Status: StatusPending, Phase: PhaseAccepted,
				Resources: []ResourceBinding{resource},
			},
		},
		{
			name: "failed",
			receipt: Receipt{
				Status: StatusFailed, Phase: PhaseFailed,
				Reason:    &Reason{Code: ReasonCodeProviderPartialCreate, Retryable: true},
				Resources: []ResourceBinding{resource},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := sealReceipt(t, command, tt.receipt, nil); !errors.Is(err, ErrInvalidReceipt) {
				t.Fatalf("plan resource injection error = %v, want ErrInvalidReceipt", err)
			}
		})
	}
}

func TestValidateForRejectsNonCanonicalReceipt(t *testing.T) {
	commandInput := validCommand(OperationObserve)
	commandInput.Targets = []ResourceTarget{resourceTarget("server"), resourceTarget("network")}
	command := mustCommand(t, commandInput)
	resources := make([]ResourceBinding, 0, len(command.Targets))
	for _, target := range command.Targets {
		resource := resourceBinding(target, ObservationPresent, CleanupRequired)
		resource.Evidence = []Evidence{
			observationEvidence(command, target, ObservationPresent, target.BindingID+"-b"),
			observationEvidence(command, target, ObservationPresent, target.BindingID+"-a"),
		}
		resources = append(resources, resource)
	}
	sealed := mustReceipt(t, command, Receipt{Status: StatusSucceeded, Phase: PhasePresent, Resources: resources}, nil)
	tests := []struct {
		name   string
		mutate func(*Receipt)
	}{
		{"whitespace", func(r *Receipt) { r.AdapterID += " " }},
		{"resource generation case", func(r *Receipt) { r.ResourceGenerationID = strings.ToUpper(r.ResourceGenerationID) }},
		{"digest case", func(r *Receipt) { r.CommandDigest = strings.ToUpper(r.CommandDigest) }},
		{"resource order", func(r *Receipt) { r.Resources[0], r.Resources[1] = r.Resources[1], r.Resources[0] }},
		{"evidence order", func(r *Receipt) {
			r.Resources[0].Evidence[0], r.Resources[0].Evidence[1] = r.Resources[0].Evidence[1], r.Resources[0].Evidence[0]
		}},
		{"evidence case", func(r *Receipt) {
			r.Resources[0].Evidence[0].Digest = strings.ToUpper(r.Resources[0].Evidence[0].Digest)
		}},
		{"timestamp location", func(r *Receipt) { r.IssuedAt = r.IssuedAt.In(time.FixedZone("review", 3600)) }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			candidate := sealed
			candidate.Resources = cloneResources(sealed.Resources)
			tt.mutate(&candidate)
			if isCanonicalReceipt(candidate) {
				t.Fatal("mutated receipt unexpectedly canonical")
			}
			if err := candidate.ValidateFor(context.Background(), command, nil); !errors.Is(err, ErrInvalidReceipt) {
				t.Fatalf("ValidateFor error = %v, want ErrInvalidReceipt", err)
			}
		})
	}

	failedCommand := mustCommand(t, validCommand(OperationPlan))
	failed := mustReceipt(t, failedCommand, Receipt{
		Status: StatusFailed,
		Phase:  PhaseFailed,
		Reason: &Reason{Code: ReasonCodeProviderTransient},
	}, nil)
	failed.Reason.Code = " " + ReasonCodeProviderTransient
	if err := failed.ValidateFor(context.Background(), failedCommand, nil); !errors.Is(err, ErrInvalidReceipt) {
		t.Fatalf("non-canonical reason error = %v, want ErrInvalidReceipt", err)
	}
}

func TestReceiptEnvelopeCanonicalization(t *testing.T) {
	command := mustCommand(t, validCommand(OperationProvision))
	first := resourceBinding(resourceTarget("server"), ObservationPresent, CleanupRequired)
	second := resourceBinding(resourceTarget("network"), ObservationPresent, CleanupRequired)
	receipt := receiptForCommand(command, Receipt{
		APIVersion: " " + APIVersion + " ",
		Status:     Status(" succeeded "),
		Phase:      Phase(" present "),
		IssuedAt:   command.RequestedAt.Add(time.Minute).In(time.FixedZone("review", 3600)),
		Resources:  []ResourceBinding{first, second},
	})
	receipt.AdapterID += " "
	receipt.CommandDigest = strings.ToUpper(receipt.CommandDigest)
	receipt.Resources[0].NativeRef += " "
	receipt.Resources[1].OwnershipHash = strings.ToUpper(receipt.Resources[1].OwnershipHash)
	sealed, err := sealReceiptEnvelope(context.Background(), command, receipt, nil)
	if err != nil {
		t.Fatalf("sealReceiptEnvelope: %v", err)
	}
	if !isCanonicalReceipt(sealed) {
		t.Fatalf("receipt envelope is non-canonical: %+v", sealed)
	}
	if sealed.Resources[0].BindingID != "network" || sealed.IssuedAt.Location() != time.UTC {
		t.Fatalf("receipt envelope did not sort resources and canonicalize time: %+v", sealed)
	}
}

func TestReceiptRejectsInvalidUTF8InEveryStringEnvelope(t *testing.T) {
	bad := string([]byte{0xff})
	command := mustCommand(t, validCommand(OperationObserve))
	resource := resourceBinding(command.Targets[0], ObservationPresent, CleanupRequired)
	resource.Evidence = []Evidence{observationEvidence(command, command.Targets[0], ObservationPresent, "utf8")}
	base := mustReceipt(t, command, Receipt{Status: StatusSucceeded, Phase: PhasePresent, Resources: []ResourceBinding{resource}}, nil)
	tests := []struct {
		name   string
		mutate func(*Receipt)
	}{
		{"api version", func(r *Receipt) { r.APIVersion = bad }},
		{"operation id", func(r *Receipt) { r.OperationID = bad }},
		{"operation", func(r *Receipt) { r.Operation = Operation(bad) }},
		{"tenant", func(r *Receipt) { r.TenantID = bad }},
		{"lease", func(r *Receipt) { r.LeaseID = bad }},
		{"runtime server", func(r *Receipt) { r.RuntimeServerID = bad }},
		{"resource generation", func(r *Receipt) { r.ResourceGenerationID = bad }},
		{"idempotency", func(r *Receipt) { r.IdempotencyKey = bad }},
		{"provider", func(r *Receipt) { r.ProviderID = bad }},
		{"adapter", func(r *Receipt) { r.AdapterID = bad }},
		{"capability snapshot hash", func(r *Receipt) { r.CapabilitySnapshotHash = bad }},
		{"command digest", func(r *Receipt) { r.CommandDigest = bad }},
		{"spec hash", func(r *Receipt) { r.DesiredSpecHash = bad }},
		{"previous digest", func(r *Receipt) { r.PreviousReceiptDigest = bad }},
		{"status", func(r *Receipt) { r.Status = Status(bad) }},
		{"phase", func(r *Receipt) { r.Phase = Phase(bad) }},
		{"receipt digest", func(r *Receipt) { r.ReceiptDigest = bad }},
		{"reason code", func(r *Receipt) { r.Reason = &Reason{Code: bad} }},
		{"resource binding", func(r *Receipt) { r.Resources[0].BindingID = bad }},
		{"resource kind", func(r *Receipt) { r.Resources[0].Kind = bad }},
		{"resource native ref", func(r *Receipt) { r.Resources[0].NativeRef = bad }},
		{"resource parent", func(r *Receipt) { r.Resources[0].ParentBindingID = bad }},
		{"resource ownership", func(r *Receipt) { r.Resources[0].OwnershipHash = bad }},
		{"resource disposition", func(r *Receipt) { r.Resources[0].Disposition = ResourceDisposition(bad) }},
		{"resource observation", func(r *Receipt) { r.Resources[0].Observation = ObservationState(bad) }},
		{"resource cleanup", func(r *Receipt) { r.Resources[0].Cleanup = CleanupState(bad) }},
		{"evidence ref", func(r *Receipt) { r.Resources[0].Evidence[0].Ref = bad }},
		{"evidence digest", func(r *Receipt) { r.Resources[0].Evidence[0].Digest = bad }},
		{"evidence source", func(r *Receipt) { r.Resources[0].Evidence[0].Source = EvidenceSource(bad) }},
		{"evidence operation", func(r *Receipt) { r.Resources[0].Evidence[0].OperationID = bad }},
		{"evidence runtime server", func(r *Receipt) { r.Resources[0].Evidence[0].RuntimeServerID = bad }},
		{"evidence provider", func(r *Receipt) { r.Resources[0].Evidence[0].ProviderID = bad }},
		{"evidence capability hash", func(r *Receipt) { r.Resources[0].Evidence[0].CapabilitySnapshotHash = bad }},
		{"evidence binding", func(r *Receipt) { r.Resources[0].Evidence[0].BindingID = bad }},
		{"evidence native hash", func(r *Receipt) { r.Resources[0].Evidence[0].NativeRefHash = bad }},
		{"evidence connection hash", func(r *Receipt) { r.Resources[0].Evidence[0].ConnectionHash = bad }},
		{"evidence profile hash", func(r *Receipt) { r.Resources[0].Evidence[0].ExecutionProfileHash = bad }},
		{"evidence graph hash", func(r *Receipt) { r.Resources[0].Evidence[0].ResourceGraphHash = bad }},
		{"evidence subject hash", func(r *Receipt) { r.Resources[0].Evidence[0].SubjectHash = bad }},
		{"evidence observation", func(r *Receipt) { r.Resources[0].Evidence[0].Observation = ObservationState(bad) }},
		{"evidence attestation ref", func(r *Receipt) { r.Resources[0].Evidence[0].AttestationRef = bad }},
		{"evidence attestation digest", func(r *Receipt) { r.Resources[0].Evidence[0].AttestationDigest = bad }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			candidate := base
			candidate.Resources = cloneResources(base.Resources)
			tt.mutate(&candidate)
			if _, err := sealReceiptEnvelope(context.Background(), command, candidate, nil); !errors.Is(err, ErrInvalidReceipt) {
				t.Fatalf("sealReceiptEnvelope error = %v, want ErrInvalidReceipt", err)
			}
		})
	}
	candidate := base
	candidate.Resources = cloneResources(base.Resources)
	candidate.Resources[0].Evidence[0].Ref = bad
	if _, err := ComputeReceiptDigest(candidate); !errors.Is(err, ErrInvalidReceipt) {
		t.Fatalf("ComputeReceiptDigest error = %v, want ErrInvalidReceipt", err)
	}
}

func TestReasonWireIsSecretFree(t *testing.T) {
	payload, err := json.Marshal(Reason{Code: ReasonCodeProviderTransient, Retryable: true})
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if string(payload) != `{"code":"provider.transient","retryable":true}` {
		t.Fatalf("reason JSON = %s", payload)
	}
}

func TestProviderReasonCodeTaxonomyIsClosed(t *testing.T) {
	want := []string{
		ReasonCodeProviderAuth,
		ReasonCodeProviderPermission,
		ReasonCodeProviderInvalidSpec,
		ReasonCodeProviderQuota,
		ReasonCodeProviderRateLimit,
		ReasonCodeProviderConflict,
		ReasonCodeProviderNotFound,
		ReasonCodeProviderTransient,
		ReasonCodeProviderTimeout,
		ReasonCodeProviderPartialCreate,
		ReasonCodeProviderCleanupRequired,
	}
	for _, code := range want {
		if !IsProviderReasonCode(code) || !validCode(code) {
			t.Errorf("canonical provider reason %q was rejected", code)
		}
	}
	for _, code := range []string{"provider.failed", "provider.partial", "provider.operation_failed"} {
		if IsProviderReasonCode(code) || validCode(code) {
			t.Errorf("non-canonical provider reason %q was accepted", code)
		}
	}
	if !validCode("policy.denied") {
		t.Error("non-provider reason namespace was unexpectedly rejected")
	}
}

func TestSuccessfulTargetedOperationsRequireExactCoverage(t *testing.T) {
	for _, operation := range []Operation{OperationObserve, OperationReconcile, OperationDecommission} {
		t.Run(string(operation), func(t *testing.T) {
			command := mustCommand(t, validCommand(operation))
			statusReceipt := Receipt{Status: StatusSucceeded, Phase: PhasePresent}
			resource := resourceBinding(command.Targets[0], ObservationPresent, CleanupRequired)
			verifier := &recordingVerifier{}
			if operation == OperationDecommission {
				statusReceipt.Phase = PhaseAbsent
				resource = absentBinding(command, command.Targets[0], "one")
			} else if operation == OperationObserve {
				resource.Evidence = []Evidence{observationEvidence(command, command.Targets[0], ObservationPresent, "one")}
			}
			statusReceipt.Resources = []ResourceBinding{resource}
			if _, err := sealReceipt(t, command, statusReceipt, verifier); err != nil {
				t.Fatalf("valid exact coverage: %v", err)
			}

			extra := statusReceipt
			extra.Resources = append(append([]ResourceBinding(nil), statusReceipt.Resources...), resourceBinding(resourceTarget("extra"), ObservationPresent, CleanupRequired))
			if operation == OperationDecommission {
				extra.Resources[1] = absentBinding(command, resourceTarget("extra"), "extra")
			}
			if _, err := sealReceipt(t, command, extra, verifier); !errors.Is(err, ErrInvalidReceipt) {
				t.Fatalf("extra coverage error = %v, want ErrInvalidReceipt", err)
			}

			mismatch := statusReceipt
			mismatch.Resources = append([]ResourceBinding(nil), statusReceipt.Resources...)
			mismatch.Resources[0].OwnershipHash = testDigest("different-owner")
			if _, err := sealReceipt(t, command, mismatch, verifier); !errors.Is(err, ErrInvalidReceipt) {
				t.Fatalf("mismatched coverage error = %v, want ErrInvalidReceipt", err)
			}
		})
	}
}

func TestTargetedIntermediateAndFailedReceiptsCannotBindForeignResources(t *testing.T) {
	for _, operation := range []Operation{OperationObserve, OperationReconcile, OperationDecommission} {
		t.Run(string(operation), func(t *testing.T) {
			input := validCommand(operation)
			input.Targets = []ResourceTarget{resourceTarget("network"), resourceTarget("server")}
			command := mustCommand(t, input)

			pending := Receipt{
				Status:    StatusPending,
				Phase:     PhaseAccepted,
				Resources: []ResourceBinding{resourceBinding(command.Targets[0], ObservationPresent, CleanupRequired)},
			}
			if _, err := sealReceipt(t, command, pending, nil); err != nil {
				t.Fatalf("pending target subset rejected: %v", err)
			}

			failed := Receipt{
				Status:    StatusFailed,
				Phase:     PhaseFailed,
				Reason:    &Reason{Code: ReasonCodeProviderPartialCreate, Retryable: true},
				Resources: []ResourceBinding{resourceBinding(command.Targets[1], ObservationPresent, CleanupRequired)},
			}
			if _, err := sealReceipt(t, command, failed, nil); err != nil {
				t.Fatalf("failed target subset rejected: %v", err)
			}

			foreign := resourceBinding(resourceTarget("foreign"), ObservationPresent, CleanupRequired)
			for _, receipt := range []Receipt{
				{Status: StatusPending, Phase: PhaseAccepted, Resources: []ResourceBinding{foreign}},
				{Status: StatusFailed, Phase: PhaseFailed, Reason: &Reason{Code: ReasonCodeProviderPartialCreate, Retryable: true}, Resources: []ResourceBinding{foreign}},
			} {
				if _, err := sealReceipt(t, command, receipt, nil); !errors.Is(err, ErrInvalidReceipt) {
					t.Fatalf("foreign resource receipt error = %v, want ErrInvalidReceipt", err)
				}
			}
		})
	}
}

func TestSuccessfulObserveUsesOneResolvedObservation(t *testing.T) {
	commandInput := validCommand(OperationObserve)
	commandInput.Targets = []ResourceTarget{resourceTarget("network"), resourceTarget("server")}
	command := mustCommand(t, commandInput)
	receipt := Receipt{
		Status: StatusSucceeded,
		Phase:  PhasePresent,
		Resources: []ResourceBinding{
			resourceBinding(command.Targets[0], ObservationPresent, CleanupRequired),
			resourceBinding(command.Targets[1], ObservationAbsent, CleanupComplete),
		},
	}
	receipt.Resources[0].Evidence = []Evidence{observationEvidence(command, command.Targets[0], ObservationPresent, "present")}
	receipt.Resources[1].Evidence = []Evidence{absenceEvidence(command, command.Targets[1], "mixed")}
	if _, err := sealReceipt(t, command, receipt, &recordingVerifier{}); !errors.Is(err, ErrInvalidReceipt) {
		t.Fatalf("mixed observe error = %v, want ErrInvalidReceipt", err)
	}
}

func TestAbsenceEvidenceIsBoundDefinitiveAttestedBoundedUniqueAndVerified(t *testing.T) {
	command := mustCommand(t, validCommand(OperationDecommission))
	valid := Receipt{
		Status:    StatusSucceeded,
		Phase:     PhaseAbsent,
		Resources: []ResourceBinding{absentBinding(command, command.Targets[0], "one")},
	}
	verifier := &recordingVerifier{}
	if _, err := sealReceipt(t, command, valid, verifier); err != nil {
		t.Fatalf("valid absence receipt: %v", err)
	}
	if verifier.calls != 1 {
		t.Fatalf("verifier calls = %d, want 1", verifier.calls)
	}
	if _, err := sealReceipt(t, command, valid, nil); !errors.Is(err, ErrAbsenceProofRequired) {
		t.Fatalf("nil verifier error = %v, want ErrAbsenceProofRequired", err)
	}
	verifier.err = errors.New("untrusted attestation")
	if _, err := sealReceipt(t, command, valid, verifier); !errors.Is(err, ErrAbsenceProofRequired) {
		t.Fatalf("verifier rejection error = %v, want ErrAbsenceProofRequired", err)
	}

	tests := []struct {
		name   string
		mutate func(*Evidence)
		want   error
	}{
		{"operation", func(e *Evidence) { e.OperationID = "other" }, ErrInvalidReceipt},
		{"binding", func(e *Evidence) { e.BindingID = "other" }, ErrInvalidReceipt},
		{"native ref", func(e *Evidence) { e.NativeRefHash = testDigest("other") }, ErrInvalidReceipt},
		{"connection", func(e *Evidence) { e.ConnectionHash = testDigest("other") }, ErrInvalidReceipt},
		{"profile", func(e *Evidence) { e.ExecutionProfileHash = testDigest("other") }, ErrInvalidReceipt},
		{"graph", func(e *Evidence) { e.ResourceGraphHash = testDigest("other") }, ErrInvalidReceipt},
		{"subject", func(e *Evidence) { e.SubjectHash = testDigest("other") }, ErrInvalidReceipt},
		{"not definitive", func(e *Evidence) { e.Definitive = false }, ErrInvalidReceipt},
		{"not attested", func(e *Evidence) { e.AttestationRef = "" }, ErrInvalidReceipt},
		{"wrong evidence scheme", func(e *Evidence) { e.Ref = "desired-spec://wrong" }, ErrInvalidReceipt},
		{"wrong attestation scheme", func(e *Evidence) { e.AttestationRef = "provider-evidence://wrong" }, ErrInvalidReceipt},
		{"too early", func(e *Evidence) { e.CollectedAt = command.RequestedAt.Add(-time.Nanosecond) }, ErrInvalidReceipt},
		{"too late", func(e *Evidence) { e.CollectedAt = command.RequestedAt.Add(2 * time.Minute) }, ErrInvalidReceipt},
		{"adapter absence", func(e *Evidence) { e.Source = EvidenceSourceAdapter }, ErrAbsenceProofRequired},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			candidate := valid
			candidate.Resources = cloneResources(valid.Resources)
			tt.mutate(&candidate.Resources[0].Evidence[0])
			if _, err := sealReceipt(t, command, candidate, &recordingVerifier{}); !errors.Is(err, tt.want) {
				t.Fatalf("error = %v, want %v", err, tt.want)
			}
		})
	}
}

func TestEvidenceMustBeUniqueAcrossResourceGraph(t *testing.T) {
	commandInput := validCommand(OperationDecommission)
	commandInput.Targets = []ResourceTarget{resourceTarget("network"), resourceTarget("server")}
	command := mustCommand(t, commandInput)
	first := absentBinding(command, command.Targets[0], "same")
	second := absentBinding(command, command.Targets[1], "other")
	second.Evidence[0].Ref = first.Evidence[0].Ref
	receipt := Receipt{Status: StatusSucceeded, Phase: PhaseAbsent, Resources: []ResourceBinding{first, second}}
	if _, err := sealReceipt(t, command, receipt, &recordingVerifier{}); !errors.Is(err, ErrInvalidReceipt) {
		t.Fatalf("duplicate evidence error = %v, want ErrInvalidReceipt", err)
	}
}

func TestResourceGraphRejectsUnknownParentsCyclesAndInvalidObservationCleanup(t *testing.T) {
	command := mustCommand(t, validCommand(OperationProvision))
	tests := []struct {
		name      string
		resources []ResourceBinding
	}{
		{"unknown parent", []ResourceBinding{func() ResourceBinding {
			r := resourceBinding(resourceTarget("server"), ObservationPresent, CleanupRequired)
			r.ParentBindingID = "missing"
			return r
		}()}},
		{"cycle", func() []ResourceBinding {
			first := resourceBinding(resourceTarget("network"), ObservationPresent, CleanupRequired)
			second := resourceBinding(resourceTarget("server"), ObservationPresent, CleanupRequired)
			first.ParentBindingID = second.BindingID
			second.ParentBindingID = first.BindingID
			return []ResourceBinding{first, second}
		}()},
		{"present complete", []ResourceBinding{resourceBinding(resourceTarget("server"), ObservationPresent, CleanupComplete)}},
		{"absent pending", []ResourceBinding{resourceBinding(resourceTarget("server"), ObservationAbsent, CleanupPending)}},
		{"unknown complete", []ResourceBinding{resourceBinding(resourceTarget("server"), ObservationUnknown, CleanupComplete)}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			receipt := Receipt{Status: StatusSucceeded, Phase: PhasePresent, Resources: tt.resources}
			if _, err := sealReceipt(t, command, receipt, &recordingVerifier{}); !errors.Is(err, ErrInvalidReceipt) {
				t.Fatalf("error = %v, want ErrInvalidReceipt", err)
			}
		})
	}
}

func TestIssuedAtMustNotPrecedeRequestedAt(t *testing.T) {
	command := mustCommand(t, validCommand(OperationPlan))
	_, err := sealReceiptAt(t, command, Receipt{Status: StatusSucceeded, Phase: PhasePlanned}, nil, command.RequestedAt.Add(-time.Nanosecond))
	if !errors.Is(err, ErrInvalidReceipt) {
		t.Fatalf("receipt envelope error = %v, want ErrInvalidReceipt", err)
	}
}

func TestReceiptSequenceAndValidateAppend(t *testing.T) {
	command := mustCommand(t, validCommand(OperationProvision))
	first := mustReceipt(t, command, Receipt{Sequence: 1, Status: StatusPending, Phase: PhaseRequested}, nil)
	secondResource := resourceBinding(resourceTarget("server"), ObservationUnknown, CleanupPending)
	second := mustReceipt(t, command, Receipt{
		Sequence:              2,
		PreviousReceiptDigest: first.ReceiptDigest,
		Status:                StatusPending,
		Phase:                 PhaseAccepted,
		IssuedAt:              first.IssuedAt.Add(time.Second),
		Resources:             []ResourceBinding{secondResource},
	}, nil)
	if err := ValidateAppend(context.Background(), command, first, second, nil); err != nil {
		t.Fatalf("ValidateAppend: %v", err)
	}
	third := mustReceipt(t, command, Receipt{
		Sequence:              3,
		PreviousReceiptDigest: second.ReceiptDigest,
		Status:                StatusPending,
		Phase:                 PhaseResourcesBound,
		IssuedAt:              second.IssuedAt.Add(time.Second),
		Resources:             []ResourceBinding{secondResource},
	}, nil)
	if err := ValidateAppend(context.Background(), command, second, third, nil); err != nil {
		t.Fatalf("ValidateAppend resources_bound: %v", err)
	}

	tests := []struct {
		name   string
		mutate func(*Receipt)
	}{
		{"sequence gap", func(r *Receipt) { r.Sequence++ }},
		{"wrong previous digest", func(r *Receipt) { r.PreviousReceiptDigest = testDigest("wrong") }},
		{"time regression", func(r *Receipt) { r.IssuedAt = first.IssuedAt.Add(-time.Nanosecond) }},
		{"invalid transition", func(r *Receipt) { r.Phase = PhasePresent; r.Status = StatusSucceeded }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			candidate := second
			candidate.Resources = cloneResources(second.Resources)
			tt.mutate(&candidate)
			candidate = resealReceipt(t, candidate)
			if err := ValidateAppend(context.Background(), command, first, candidate, nil); !errors.Is(err, ErrInvalidReceipt) {
				t.Fatalf("ValidateAppend error = %v, want ErrInvalidReceipt", err)
			}
		})
	}

	failure := mustReceipt(t, command, Receipt{
		Sequence:              4,
		PreviousReceiptDigest: third.ReceiptDigest,
		Status:                StatusFailed,
		Phase:                 PhaseFailed,
		IssuedAt:              third.IssuedAt.Add(time.Second),
		Reason:                &Reason{Code: ReasonCodeProviderTransient},
		Resources:             []ResourceBinding{secondResource},
	}, nil)
	if err := ValidateAppend(context.Background(), command, third, failure, nil); err != nil {
		t.Fatalf("ValidateAppend failure with retained graph: %v", err)
	}
	for _, tt := range []struct {
		name   string
		mutate func(*Receipt)
	}{
		{"dropped graph binding", func(r *Receipt) { r.Resources = nil }},
		{"mutated graph identity", func(r *Receipt) { r.Resources[0].NativeRef = "mutated" }},
	} {
		t.Run(tt.name, func(t *testing.T) {
			candidate := failure
			candidate.Resources = cloneResources(failure.Resources)
			tt.mutate(&candidate)
			candidate = resealReceipt(t, candidate)
			if err := ValidateAppend(context.Background(), command, third, candidate, nil); !errors.Is(err, ErrInvalidReceipt) {
				t.Fatalf("ValidateAppend error = %v, want ErrInvalidReceipt", err)
			}
		})
	}
}

func TestValidateAppendRetainsCompleteEvidenceEnvelope(t *testing.T) {
	command := mustCommand(t, validCommand(OperationDecommission))
	resource := resourceBinding(command.Targets[0], ObservationPresent, CleanupPending)
	resource.Evidence = []Evidence{observationEvidence(command, command.Targets[0], ObservationPresent, "continuity")}
	previous := mustReceipt(t, command, Receipt{
		Status:    StatusPending,
		Phase:     PhaseAccepted,
		Resources: []ResourceBinding{resource},
	}, nil)
	next := mustReceipt(t, command, Receipt{
		Sequence:              3,
		PreviousReceiptDigest: previous.ReceiptDigest,
		Status:                StatusPending,
		Phase:                 PhaseDeleteAccepted,
		IssuedAt:              previous.IssuedAt.Add(time.Second),
		Resources:             []ResourceBinding{resource},
	}, nil)
	if err := ValidateAppend(context.Background(), command, previous, next, nil); err != nil {
		t.Fatalf("ValidateAppend unchanged evidence: %v", err)
	}

	mutated := next
	mutated.Resources = cloneResources(next.Resources)
	mutated.Resources[0].Evidence[0].AttestationDigest = testDigest("different-attestation")
	mutated = resealReceipt(t, mutated)
	if mutated.Resources[0].Evidence[0].Digest != previous.Resources[0].Evidence[0].Digest {
		t.Fatal("test unexpectedly changed the payload digest")
	}
	if err := ValidateAppend(context.Background(), command, previous, mutated, nil); !errors.Is(err, ErrInvalidReceipt) {
		t.Fatalf("ValidateAppend error = %v, want ErrInvalidReceipt", err)
	}
}

func TestReceiptSequenceEnvelopeRules(t *testing.T) {
	command := mustCommand(t, validCommand(OperationPlan))
	valid := mustReceipt(t, command, Receipt{Status: StatusSucceeded, Phase: PhasePlanned}, nil)
	valid.Sequence = 0
	valid = resealReceipt(t, valid)
	if err := valid.ValidateFor(context.Background(), command, nil); !errors.Is(err, ErrInvalidReceipt) {
		t.Fatalf("zero sequence error = %v, want ErrInvalidReceipt", err)
	}
	if _, err := sealReceipt(t, command, Receipt{Sequence: 1, PreviousReceiptDigest: testDigest("prior"), Status: StatusSucceeded, Phase: PhasePlanned}, nil); !errors.Is(err, ErrInvalidReceipt) {
		t.Fatalf("first previous digest error = %v, want ErrInvalidReceipt", err)
	}
	if _, err := sealReceipt(t, command, Receipt{Sequence: 2, Status: StatusSucceeded, Phase: PhasePlanned}, nil); !errors.Is(err, ErrInvalidReceipt) {
		t.Fatalf("later missing previous digest error = %v, want ErrInvalidReceipt", err)
	}
	if _, err := sealReceipt(t, command, Receipt{Sequence: 1, Status: StatusSucceeded, Phase: PhasePlanned}, nil); !errors.Is(err, ErrInvalidReceipt) {
		t.Fatalf("non-initial first receipt error = %v, want ErrInvalidReceipt", err)
	}
	if _, err := sealReceipt(t, command, Receipt{Sequence: 1, Status: StatusPending, Phase: PhaseRequested, Resources: []ResourceBinding{resourceBinding(resourceTarget("unexpected"), ObservationUnknown, CleanupPending)}}, nil); !errors.Is(err, ErrInvalidReceipt) {
		t.Fatalf("first receipt resources error = %v, want ErrInvalidReceipt", err)
	}
	if _, err := sealReceipt(t, command, Receipt{Sequence: 2, PreviousReceiptDigest: testDigest("initial"), Status: StatusPending, Phase: PhaseRequested}, nil); !errors.Is(err, ErrInvalidReceipt) {
		t.Fatalf("later requested receipt error = %v, want ErrInvalidReceipt", err)
	}
}

func TestCoordinatorOwnsReceiptAssemblyAndChainHead(t *testing.T) {
	command := mustCommand(t, validCommand(OperationPlan))
	first, err := InitialReceipt(command, command.RequestedAt.Add(time.Second))
	if err != nil {
		t.Fatalf("InitialReceipt: %v", err)
	}
	if first.Sequence != 1 || first.Status != StatusPending || first.Phase != PhaseRequested || len(first.Resources) != 0 {
		t.Fatalf("unexpected initial receipt: %+v", first)
	}
	next, err := AssembleReceipt(context.Background(), ExecutionRequest{Command: command, Previous: first}, ExecutionResult{
		Status: StatusPending,
		Phase:  PhaseAccepted,
	}, first.IssuedAt.Add(time.Second), nil)
	if err != nil {
		t.Fatalf("AssembleReceipt: %v", err)
	}
	if next.Sequence != 2 || next.PreviousReceiptDigest != first.ReceiptDigest {
		t.Fatalf("coordinator did not assemble next chain entry: %+v", next)
	}

	decommission := mustCommand(t, validCommand(OperationDecommission))
	initial, err := InitialReceipt(decommission, decommission.RequestedAt.Add(time.Second))
	if err != nil {
		t.Fatalf("InitialReceipt decommission: %v", err)
	}
	result := ExecutionResult{
		Status:    StatusSucceeded,
		Phase:     PhaseAbsent,
		Resources: []ResourceBinding{absentBinding(decommission, decommission.Targets[0], "jump")},
	}
	if _, err := AssembleReceipt(context.Background(), ExecutionRequest{Command: decommission, Previous: initial}, result, initial.IssuedAt.Add(time.Second), &recordingVerifier{}); !errors.Is(err, ErrInvalidReceipt) {
		t.Fatalf("coordinator accepted requested -> absent jump: %v", err)
	}
}

func TestAdapterResultsCannotDenyAndMustUseProviderReasons(t *testing.T) {
	command := mustCommand(t, validCommand(OperationPlan))
	initial, err := InitialReceipt(command, command.RequestedAt)
	if err != nil {
		t.Fatalf("InitialReceipt: %v", err)
	}
	request := ExecutionRequest{Command: command, Previous: initial}
	for _, tt := range []struct {
		name   string
		result ExecutionResult
	}{
		{"denied", ExecutionResult{Status: StatusDenied, Phase: PhaseDenied, Reason: &Reason{Code: "policy.denied"}}},
		{"non-provider failure", ExecutionResult{Status: StatusFailed, Phase: PhaseFailed, Reason: &Reason{Code: "coordinator.failed"}}},
		{"unknown provider failure", ExecutionResult{Status: StatusFailed, Phase: PhaseFailed, Reason: &Reason{Code: "provider.vendor_specific"}}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := AssembleReceipt(context.Background(), request, tt.result, initial.IssuedAt.Add(time.Second), nil); !errors.Is(err, ErrInvalidReceipt) {
				t.Fatalf("AssembleReceipt error = %v, want ErrInvalidReceipt", err)
			}
		})
	}

	failed, err := AssembleReceipt(context.Background(), request, ExecutionResult{
		Status: StatusFailed,
		Phase:  PhaseFailed,
		Reason: &Reason{Code: ReasonCodeProviderTransient, Retryable: true},
	}, initial.IssuedAt.Add(time.Second), nil)
	if err != nil {
		t.Fatalf("canonical provider failure: %v", err)
	}
	if failed.Reason == nil || failed.Reason.Code != ReasonCodeProviderTransient {
		t.Fatalf("provider failure reason lost: %+v", failed)
	}
}

func TestAdmissionDenialIsRequestedOnly(t *testing.T) {
	command := mustCommand(t, validCommand(OperationPlan))
	initial, err := InitialReceipt(command, command.RequestedAt)
	if err != nil {
		t.Fatalf("InitialReceipt: %v", err)
	}
	denied, err := DenyReceipt(context.Background(), ExecutionRequest{Command: command, Previous: initial}, Reason{Code: "policy.denied"}, initial.IssuedAt.Add(time.Second), nil)
	if err != nil {
		t.Fatalf("DenyReceipt: %v", err)
	}
	if denied.Status != StatusDenied || denied.Phase != PhaseDenied || len(denied.Resources) != 0 {
		t.Fatalf("unexpected denial: %+v", denied)
	}
	if _, err := DenyReceipt(context.Background(), ExecutionRequest{Command: command, Previous: initial}, Reason{Code: ReasonCodeProviderAuth}, initial.IssuedAt.Add(time.Second), nil); !errors.Is(err, ErrInvalidReceipt) {
		t.Fatalf("provider-coded denial error = %v, want ErrInvalidReceipt", err)
	}

	accepted, err := AssembleReceipt(context.Background(), ExecutionRequest{Command: command, Previous: initial}, ExecutionResult{
		Status: StatusPending,
		Phase:  PhaseAccepted,
	}, initial.IssuedAt.Add(time.Second), nil)
	if err != nil {
		t.Fatalf("accepted receipt: %v", err)
	}
	if _, err := DenyReceipt(context.Background(), ExecutionRequest{Command: command, Previous: accepted}, Reason{Code: "policy.denied"}, accepted.IssuedAt.Add(time.Second), nil); !errors.Is(err, ErrInvalidReceipt) {
		t.Fatalf("post-accept denial error = %v, want ErrInvalidReceipt", err)
	}
	if CanTransition(OperationPlan, PhaseAccepted, PhaseDenied) {
		t.Fatal("accepted transitioned to denied after executor custody began")
	}
}

func TestEvidenceVerificationRejectsTypedNilAndCanceledContext(t *testing.T) {
	command := mustCommand(t, validCommand(OperationObserve))
	resource := absentBinding(command, command.Targets[0], "typed-nil")
	receipt := Receipt{Status: StatusSucceeded, Phase: PhaseAbsent, Resources: []ResourceBinding{resource}}
	var verifier *recordingVerifier
	if _, err := sealReceipt(t, command, receipt, verifier); !errors.Is(err, ErrAbsenceProofRequired) {
		t.Fatalf("typed-nil verifier error = %v, want ErrAbsenceProofRequired", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	candidate := receiptForCommand(command, receipt)
	if _, err := sealReceiptEnvelope(ctx, command, candidate, &recordingVerifier{}); !errors.Is(err, context.Canceled) || !errors.Is(err, ErrInvalidReceipt) {
		t.Fatalf("canceled context error = %v, want context.Canceled and ErrInvalidReceipt", err)
	}
}

func TestDecommissionAppendChainPreservesObservationHistory(t *testing.T) {
	command := mustCommand(t, validCommand(OperationDecommission))
	verifier := &recordingVerifier{}
	first := mustReceipt(t, command, Receipt{
		Sequence: 1,
		Status:   StatusPending,
		Phase:    PhaseRequested,
	}, verifier)

	present := resourceBinding(command.Targets[0], ObservationPresent, CleanupRequired)
	present.Evidence = []Evidence{observationEvidence(command, command.Targets[0], ObservationPresent, "initial-present")}
	accepted := mustReceipt(t, command, Receipt{
		Sequence:              2,
		PreviousReceiptDigest: first.ReceiptDigest,
		Status:                StatusPending,
		Phase:                 PhaseAccepted,
		IssuedAt:              first.IssuedAt.Add(time.Second),
		Resources:             []ResourceBinding{present},
	}, verifier)
	if err := ValidateAppend(context.Background(), command, first, accepted, verifier); err != nil {
		t.Fatalf("append accepted: %v", err)
	}

	deleteAcceptedResource := cloneResources(accepted.Resources)[0]
	deleteAcceptedResource.Cleanup = CleanupPending
	deleteAccepted := mustReceipt(t, command, Receipt{
		Sequence:              3,
		PreviousReceiptDigest: accepted.ReceiptDigest,
		Status:                StatusPending,
		Phase:                 PhaseDeleteAccepted,
		IssuedAt:              accepted.IssuedAt.Add(time.Second),
		Resources:             []ResourceBinding{deleteAcceptedResource},
	}, verifier)
	if err := ValidateAppend(context.Background(), command, accepted, deleteAccepted, verifier); err != nil {
		t.Fatalf("append delete accepted: %v", err)
	}

	unknown := cloneResources(deleteAccepted.Resources)[0]
	unknown.Observation = ObservationUnknown
	unknown.Evidence = append(unknown.Evidence, observationEvidence(command, command.Targets[0], ObservationUnknown, "absence-pending"))
	unknown.Evidence[len(unknown.Evidence)-1].CollectedAt = command.RequestedAt.Add(31 * time.Second)
	absencePending := mustReceipt(t, command, Receipt{
		Sequence:              4,
		PreviousReceiptDigest: deleteAccepted.ReceiptDigest,
		Status:                StatusPending,
		Phase:                 PhaseAbsencePending,
		IssuedAt:              deleteAccepted.IssuedAt.Add(time.Second),
		Resources:             []ResourceBinding{unknown},
	}, verifier)
	if err := ValidateAppend(context.Background(), command, deleteAccepted, absencePending, verifier); err != nil {
		t.Fatalf("append absence pending with retained present evidence: %v", err)
	}

	absent := cloneResources(absencePending.Resources)[0]
	absent.Observation = ObservationAbsent
	absent.Cleanup = CleanupComplete
	absent.Evidence = append(absent.Evidence, absenceEvidence(command, command.Targets[0], "final-absence"))
	absent.Evidence[len(absent.Evidence)-1].CollectedAt = absencePending.PhaseEnteredAt.Add(500 * time.Millisecond)
	completed := mustReceipt(t, command, Receipt{
		Sequence:              5,
		PreviousReceiptDigest: absencePending.ReceiptDigest,
		Status:                StatusSucceeded,
		Phase:                 PhaseAbsent,
		IssuedAt:              absencePending.IssuedAt.Add(time.Second),
		Resources:             []ResourceBinding{absent},
	}, verifier)
	if err := ValidateAppend(context.Background(), command, absencePending, completed, verifier); err != nil {
		t.Fatalf("append definitive absence with preserved history: %v", err)
	}
	observations := map[ObservationState]bool{}
	for _, evidence := range completed.Resources[0].Evidence {
		observations[evidence.Observation] = true
	}
	if len(completed.Resources[0].Evidence) != 3 || !observations[ObservationPresent] || !observations[ObservationUnknown] || !observations[ObservationAbsent] {
		t.Fatalf("canonical evidence history was not retained: %+v", completed.Resources[0].Evidence)
	}
	if verifier.calls < 2 {
		t.Fatalf("absence verifier calls = %d, want at least 2", verifier.calls)
	}

	jump := completed
	jump.Sequence = 2
	jump.PreviousReceiptDigest = first.ReceiptDigest
	jump.IssuedAt = first.IssuedAt.Add(time.Second)
	jump = resealReceipt(t, jump)
	if err := ValidateAppend(context.Background(), command, first, jump, verifier); !errors.Is(err, ErrInvalidReceipt) {
		t.Fatalf("decommission requested -> absent error = %v, want ErrInvalidReceipt", err)
	}
}

func TestDecommissionAbsenceEvidenceMustFollowAbsencePendingFence(t *testing.T) {
	command := mustCommand(t, validCommand(OperationDecommission))
	verifier := &recordingVerifier{}
	previousResource := resourceBinding(command.Targets[0], ObservationUnknown, CleanupPending)
	previous := mustReceipt(t, command, Receipt{
		Status:    StatusPending,
		Phase:     PhaseAbsencePending,
		Resources: []ResourceBinding{previousResource},
	}, verifier)

	stale := absentBinding(command, command.Targets[0], "stale-after-delete")
	stale.Evidence[0].CollectedAt = previous.PhaseEnteredAt.Add(-time.Second)
	next := mustReceipt(t, command, Receipt{
		Sequence:              previous.Sequence + 1,
		PreviousReceiptDigest: previous.ReceiptDigest,
		Status:                StatusSucceeded,
		Phase:                 PhaseAbsent,
		IssuedAt:              previous.IssuedAt.Add(time.Second),
		Resources:             []ResourceBinding{stale},
	}, verifier)
	if err := ValidateAppend(context.Background(), command, previous, next, verifier); !errors.Is(err, ErrAbsenceProofRequired) {
		t.Fatalf("stale absence error = %v, want ErrAbsenceProofRequired", err)
	}

	fresh := cloneResources(next.Resources)
	fresh[0].Evidence[0].CollectedAt = previous.PhaseEnteredAt
	next.Resources = fresh
	next = resealReceipt(t, next)
	if err := ValidateAppend(context.Background(), command, previous, next, verifier); err != nil {
		t.Fatalf("fresh absence rejected: %v", err)
	}
}

func TestLatestDefinitiveEvidenceDeterminesCurrentObservation(t *testing.T) {
	input := validCommand(OperationObserve)
	input.Targets = []ResourceTarget{resourceTarget("server")}
	command := mustCommand(t, input)
	verifier := &recordingVerifier{}
	resource := resourceBinding(command.Targets[0], ObservationPresent, CleanupRequired)
	staleAbsence := absenceEvidence(command, command.Targets[0], "stale-absence")
	staleAbsence.CollectedAt = command.RequestedAt.Add(30 * time.Second)
	newerPresence := observationEvidence(command, command.Targets[0], ObservationPresent, "newer-presence")
	newerPresence.CollectedAt = command.RequestedAt.Add(31 * time.Second)
	resource.Evidence = []Evidence{staleAbsence, newerPresence}
	if _, err := sealReceipt(t, command, Receipt{Status: StatusSucceeded, Phase: PhasePresent, Resources: []ResourceBinding{resource}}, verifier); err != nil {
		t.Fatalf("newer present evidence must outrank stale absence: %v", err)
	}

	resource.Observation = ObservationAbsent
	resource.Cleanup = CleanupComplete
	if _, err := sealReceipt(t, command, Receipt{Status: StatusSucceeded, Phase: PhaseAbsent, Resources: []ResourceBinding{resource}}, verifier); !errors.Is(err, ErrInvalidReceipt) {
		t.Fatalf("stale absence must not outrank newer presence: %v", err)
	}
}

func TestProviderExecutorEnvelopeBounds(t *testing.T) {
	t.Run("targets", func(t *testing.T) {
		command := validCommand(OperationObserve)
		command.Targets = targetsForTest(MaxTargets + 1)
		if _, err := SealCommand(command); !errors.Is(err, ErrInvalidCommand) {
			t.Fatalf("too many targets error = %v, want ErrInvalidCommand", err)
		}
	})

	t.Run("resources", func(t *testing.T) {
		command := mustCommand(t, validCommand(OperationProvision))
		resources := make([]ResourceBinding, 0, MaxResources+1)
		for i := 0; i <= MaxResources; i++ {
			resources = append(resources, resourceBinding(resourceTarget(fmt.Sprintf("resource-%03d", i)), ObservationPresent, CleanupRequired))
		}
		if _, err := sealReceipt(t, command, Receipt{Status: StatusSucceeded, Phase: PhasePresent, Resources: resources}, nil); !errors.Is(err, ErrInvalidReceipt) {
			t.Fatalf("too many resources error = %v, want ErrInvalidReceipt", err)
		}
	})

	t.Run("per resource evidence", func(t *testing.T) {
		command := mustCommand(t, validCommand(OperationProvision))
		resource := resourceBinding(resourceTarget("server"), ObservationPresent, CleanupRequired)
		for i := 0; i <= MaxEvidencePerResource; i++ {
			resource.Evidence = append(resource.Evidence, observationEvidence(command, resourceTarget("server"), ObservationPresent, fmt.Sprintf("per-resource-%03d", i)))
		}
		if _, err := sealReceipt(t, command, Receipt{Status: StatusSucceeded, Phase: PhasePresent, Resources: []ResourceBinding{resource}}, nil); !errors.Is(err, ErrInvalidReceipt) {
			t.Fatalf("too much per-resource evidence error = %v, want ErrInvalidReceipt", err)
		}
	})

	t.Run("total evidence", func(t *testing.T) {
		command := mustCommand(t, validCommand(OperationProvision))
		resources := make([]ResourceBinding, 0, MaxEvidencePerReceipt/MaxEvidencePerResource+1)
		for resourceIndex := 0; len(resources)*MaxEvidencePerResource <= MaxEvidencePerReceipt; resourceIndex++ {
			target := resourceTarget(fmt.Sprintf("resource-%03d", resourceIndex))
			resource := resourceBinding(target, ObservationPresent, CleanupRequired)
			for evidenceIndex := 0; evidenceIndex < MaxEvidencePerResource; evidenceIndex++ {
				resource.Evidence = append(resource.Evidence, observationEvidence(command, target, ObservationPresent, fmt.Sprintf("total-%03d-%03d", resourceIndex, evidenceIndex)))
			}
			resources = append(resources, resource)
		}
		if len(resources)*MaxEvidencePerResource <= MaxEvidencePerReceipt {
			t.Fatal("test did not exceed total evidence limit")
		}
		if _, err := sealReceipt(t, command, Receipt{Status: StatusSucceeded, Phase: PhasePresent, Resources: resources}, nil); !errors.Is(err, ErrInvalidReceipt) {
			t.Fatalf("too much total evidence error = %v, want ErrInvalidReceipt", err)
		}
	})
}

func TestPreflightBoundsAndJSONSafeIntegersFailClosed(t *testing.T) {
	oversized := validCommand(OperationObserve)
	oversized.TenantID = strings.Repeat("t", maxIdentifierBytes+1)
	if _, err := SealCommand(oversized); !errors.Is(err, ErrInvalidCommand) {
		t.Fatalf("oversized command error = %v, want ErrInvalidCommand", err)
	}
	if _, err := ComputeCommandDigest(oversized); !errors.Is(err, ErrInvalidCommand) {
		t.Fatalf("oversized digest input error = %v, want ErrInvalidCommand", err)
	}
	if _, err := ComputeResourceGraphHash(targetsForTest(MaxTargets + 1)); !errors.Is(err, ErrInvalidCommand) {
		t.Fatalf("oversized graph error = %v, want ErrInvalidCommand", err)
	}

	for _, tt := range []struct {
		name   string
		mutate func(*Command)
	}{
		{"lease revision", func(c *Command) { c.LeaseRevision = MaxJSONSafeInteger + 1 }},
		{"ledger revision", func(c *Command) { c.LedgerRevision = MaxJSONSafeInteger + 1 }},
	} {
		t.Run(tt.name, func(t *testing.T) {
			command := validCommand(OperationPlan)
			tt.mutate(&command)
			if _, err := SealCommand(command); !errors.Is(err, ErrInvalidCommand) {
				t.Fatalf("SealCommand error = %v, want ErrInvalidCommand", err)
			}
		})
	}

	command := mustCommand(t, validCommand(OperationPlan))
	receipt := receiptForCommand(command, Receipt{Status: StatusPending, Phase: PhaseAccepted})
	receipt.Sequence = MaxJSONSafeInteger + 1
	receipt.ReceiptDigest = ""
	digest, err := ComputeReceiptDigest(receipt)
	if err != nil {
		t.Fatalf("ComputeReceiptDigest: %v", err)
	}
	receipt.ReceiptDigest = digest
	if err := receipt.ValidateFor(context.Background(), command, nil); !errors.Is(err, ErrInvalidReceipt) {
		t.Fatalf("unsafe sequence error = %v, want ErrInvalidReceipt", err)
	}

	maxHead := receiptForCommand(command, Receipt{Status: StatusPending, Phase: PhaseAccepted})
	maxHead.Sequence = MaxJSONSafeInteger
	maxHead = resealReceipt(t, maxHead)
	if _, err := AssembleReceipt(context.Background(), ExecutionRequest{Command: command, Previous: maxHead}, ExecutionResult{
		Status: StatusPending,
		Phase:  PhaseAccepted,
	}, maxHead.IssuedAt.Add(time.Second), nil); !errors.Is(err, ErrInvalidReceipt) {
		t.Fatalf("sequence exhaustion error = %v, want ErrInvalidReceipt", err)
	}
}

func TestParentGraphTraversalIsBoundedAndIterative(t *testing.T) {
	command := validCommand(OperationObserve)
	command.Targets = targetsForTest(MaxTargets)
	for i := 1; i < len(command.Targets); i++ {
		command.Targets[i].ParentBindingID = command.Targets[i-1].BindingID
	}
	if _, err := SealCommand(command); err != nil {
		t.Fatalf("maximum-size acyclic graph: %v", err)
	}
	command.Targets[0].ParentBindingID = command.Targets[len(command.Targets)-1].BindingID
	if _, err := SealCommand(command); !errors.Is(err, ErrInvalidCommand) {
		t.Fatalf("maximum-size cycle error = %v, want ErrInvalidCommand", err)
	}
}

func TestIdentifiersAndOpaqueRefsRejectRawControlAndFormatCharacters(t *testing.T) {
	for _, tt := range []struct {
		name   string
		mutate func(*Command)
	}{
		{"identifier control", func(c *Command) { c.TenantID = "tenant\x00one" }},
		{"identifier format", func(c *Command) { c.LeaseID = "lease\u200bone" }},
		{"native ref format", func(c *Command) { c.Targets[0].NativeRef = "server\u2060one" }},
		{"opaque ref control", func(c *Command) { c.CustodyRef = "custody://tenant\x00one" }},
		{"opaque ref format", func(c *Command) { c.ConnectionRef = "provider-connection://tenant\u200bone" }},
	} {
		t.Run(tt.name, func(t *testing.T) {
			command := validCommand(OperationObserve)
			tt.mutate(&command)
			if _, err := SealCommand(command); !errors.Is(err, ErrInvalidCommand) {
				t.Fatalf("SealCommand error = %v, want ErrInvalidCommand", err)
			}
		})
	}

	command := mustCommand(t, validCommand(OperationProvision))
	resource := resourceBinding(resourceTarget("server"), ObservationPresent, CleanupRequired)
	resource.NativeRef += "\u200b"
	if _, err := sealReceipt(t, command, Receipt{Status: StatusSucceeded, Phase: PhasePresent, Resources: []ResourceBinding{resource}}, nil); !errors.Is(err, ErrInvalidReceipt) {
		t.Fatalf("resource native ref format error = %v, want ErrInvalidReceipt", err)
	}
	resource = resourceBinding(resourceTarget("server"), ObservationPresent, CleanupRequired)
	resource.Evidence = []Evidence{observationEvidence(command, resourceTarget("server"), ObservationPresent, "format-ref")}
	resource.Evidence[0].Ref += "\u200b"
	if _, err := sealReceipt(t, command, Receipt{Status: StatusSucceeded, Phase: PhasePresent, Resources: []ResourceBinding{resource}}, nil); !errors.Is(err, ErrInvalidReceipt) {
		t.Fatalf("evidence opaque ref format error = %v, want ErrInvalidReceipt", err)
	}
}

func TestFailedReceiptMayRetainPartialResources(t *testing.T) {
	command := mustCommand(t, validCommand(OperationProvision))
	partial := resourceBinding(resourceTarget("partial-server"), ObservationPresent, CleanupRequired)
	receipt, err := sealReceipt(t, command, Receipt{
		Status:    StatusFailed,
		Phase:     PhaseFailed,
		Reason:    &Reason{Code: ReasonCodeProviderPartialCreate, Retryable: true},
		Resources: []ResourceBinding{partial},
	}, nil)
	if err != nil {
		t.Fatalf("failed partial receipt: %v", err)
	}
	if len(receipt.Resources) != 1 || receipt.Resources[0].NativeRef == "" {
		t.Fatalf("partial cleanup handle not retained: %+v", receipt.Resources)
	}
}

func TestFailedReceiptRequiresReasonEnvelope(t *testing.T) {
	command := mustCommand(t, validCommand(OperationObserve))
	_, err := sealReceipt(t, command, Receipt{Status: StatusFailed, Phase: PhaseFailed}, nil)
	if !errors.Is(err, ErrInvalidReceipt) {
		t.Fatalf("receipt envelope error = %v, want ErrInvalidReceipt", err)
	}
}

func TestDeniedReceiptCannotCarrySideEffectCustody(t *testing.T) {
	command := mustCommand(t, validCommand(OperationProvision))
	resource := resourceBinding(resourceTarget("partial-server"), ObservationUnknown, CleanupRequired)
	_, err := sealReceipt(t, command, Receipt{
		Status: StatusDenied,
		Phase:  PhaseDenied,
		Reason: &Reason{Code: "policy.denied"},
		Resources: []ResourceBinding{
			resource,
		},
	}, nil)
	if !errors.Is(err, ErrInvalidReceipt) {
		t.Fatalf("denied receipt with resource error = %v, want ErrInvalidReceipt", err)
	}
	if CanTransition(OperationProvision, PhaseResourcesBound, PhaseDenied) {
		t.Fatal("resources_bound transitioned to denied instead of cleanup-preserving failed")
	}
}

func TestLifecycleTransitionsDoNotSkipAbsencePending(t *testing.T) {
	steps := [][2]Phase{
		{PhaseRequested, PhaseAccepted},
		{PhaseAccepted, PhaseDeleteAccepted},
		{PhaseDeleteAccepted, PhaseAbsencePending},
		{PhaseAbsencePending, PhaseAbsent},
	}
	for _, step := range steps {
		if err := ValidateTransition(OperationDecommission, step[0], step[1]); err != nil {
			t.Fatalf("ValidateTransition(%s, %s): %v", step[0], step[1], err)
		}
	}
	if CanTransition(OperationDecommission, PhaseDeleteAccepted, PhaseAbsent) {
		t.Fatal("decommission skipped absence_pending")
	}
}

func TestPendingSelfLoopsRetainAsyncGraphAndEvidence(t *testing.T) {
	tests := []struct {
		name        string
		operation   Operation
		phase       Phase
		observation ObservationState
		cleanup     CleanupState
		resource    bool
	}{
		{name: "plan-accepted", operation: OperationPlan, phase: PhaseAccepted},
		{name: "observe-accepted", operation: OperationObserve, phase: PhaseAccepted, observation: ObservationPresent, cleanup: CleanupRequired, resource: true},
		{name: "provision-resources-bound", operation: OperationProvision, phase: PhaseResourcesBound, observation: ObservationPresent, cleanup: CleanupPending, resource: true},
		{name: "reconcile-resources-bound", operation: OperationReconcile, phase: PhaseResourcesBound, observation: ObservationPresent, cleanup: CleanupPending, resource: true},
		{name: "decommission-delete-accepted", operation: OperationDecommission, phase: PhaseDeleteAccepted, observation: ObservationPresent, cleanup: CleanupPending, resource: true},
		{name: "decommission-absence-pending", operation: OperationDecommission, phase: PhaseAbsencePending, observation: ObservationUnknown, cleanup: CleanupPending, resource: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			command := mustCommand(t, validCommand(tt.operation))
			var resources []ResourceBinding
			var target ResourceTarget
			if tt.resource {
				if targetBoundOperation(tt.operation) {
					target = command.Targets[0]
				} else {
					target = resourceTarget("created-server")
				}
				resource := resourceBinding(target, tt.observation, tt.cleanup)
				resource.Evidence = []Evidence{observationEvidence(command, target, tt.observation, tt.name+"-initial")}
				resources = []ResourceBinding{resource}
			}

			previous := mustReceipt(t, command, Receipt{
				Status: StatusPending, Phase: tt.phase, Resources: resources,
			}, nil)
			nextResources := cloneResources(previous.Resources)
			if tt.resource {
				evidence := observationEvidence(command, target, tt.observation, tt.name+"-next")
				evidence.CollectedAt = command.RequestedAt.Add(31 * time.Second)
				nextResources[0].Evidence = append(nextResources[0].Evidence, evidence)
			}
			next := mustReceipt(t, command, Receipt{
				Sequence: previous.Sequence + 1, PreviousReceiptDigest: previous.ReceiptDigest,
				Status: StatusPending, Phase: tt.phase, Resources: nextResources,
				PhaseEnteredAt: previous.PhaseEnteredAt,
				IssuedAt:       previous.IssuedAt.Add(time.Second),
			}, nil)
			if err := ValidateAppend(context.Background(), command, previous, next, nil); err != nil {
				t.Fatalf("pending self-loop rejected: %v", err)
			}
			if tt.resource && len(next.Resources[0].Evidence) != len(previous.Resources[0].Evidence)+1 {
				t.Fatalf("self-loop evidence extension lost: previous=%+v next=%+v", previous.Resources, next.Resources)
			}
		})
	}
}

func TestSelfLoopsRejectRequestedTerminalAndNonPendingStates(t *testing.T) {
	for _, tt := range []struct {
		name      string
		operation Operation
		phase     Phase
	}{
		{name: "requested", operation: OperationProvision, phase: PhaseRequested},
		{name: "planned", operation: OperationPlan, phase: PhasePlanned},
		{name: "present", operation: OperationProvision, phase: PhasePresent},
		{name: "absent", operation: OperationObserve, phase: PhaseAbsent},
		{name: "failed", operation: OperationProvision, phase: PhaseFailed},
		{name: "denied", operation: OperationProvision, phase: PhaseDenied},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if CanTransition(tt.operation, tt.phase, tt.phase) {
				t.Fatalf("CanTransition accepted %s self-loop", tt.phase)
			}
		})
	}

	command := mustCommand(t, validCommand(OperationPlan))
	previous := mustReceipt(t, command, Receipt{Status: StatusPending, Phase: PhaseAccepted}, nil)
	for _, tt := range []struct {
		name   string
		status Status
		reason *Reason
	}{
		{name: "succeeded", status: StatusSucceeded},
		{name: "failed", status: StatusFailed, reason: &Reason{Code: ReasonCodeProviderTransient}},
		{name: "denied", status: StatusDenied, reason: &Reason{Code: "policy.denied"}},
	} {
		t.Run("accepted-"+tt.name, func(t *testing.T) {
			candidate := receiptForCommand(command, Receipt{
				Sequence: previous.Sequence + 1, PreviousReceiptDigest: previous.ReceiptDigest,
				Status: tt.status, Phase: PhaseAccepted, Reason: tt.reason,
				IssuedAt: previous.IssuedAt.Add(time.Second),
			})
			candidate = resealReceipt(t, candidate)
			if err := ValidateAppend(context.Background(), command, previous, candidate, nil); !errors.Is(err, ErrInvalidReceipt) {
				t.Fatalf("non-pending self-loop error = %v, want ErrInvalidReceipt", err)
			}
		})
	}

	terminal := mustReceipt(t, command, Receipt{Status: StatusSucceeded, Phase: PhasePlanned}, nil)
	terminalLoop := mustReceipt(t, command, Receipt{
		Sequence: terminal.Sequence + 1, PreviousReceiptDigest: terminal.ReceiptDigest,
		Status: StatusSucceeded, Phase: PhasePlanned,
		IssuedAt: terminal.IssuedAt.Add(time.Second),
	}, nil)
	if err := ValidateAppend(context.Background(), command, terminal, terminalLoop, nil); !errors.Is(err, ErrInvalidReceipt) {
		t.Fatalf("terminal self-loop error = %v, want ErrInvalidReceipt", err)
	}
}

func TestDigestHelpersAndReceiptMutation(t *testing.T) {
	command := mustCommand(t, validCommand(OperationPlan))
	if ComputeOperationID(command.TenantID, command.LeaseID, command.ResourceGenerationID, command.Operation, command.IdempotencyKey) != command.OperationID {
		t.Fatal("ComputeOperationID mismatch")
	}
	if digest, err := ComputeCommandDigest(command); err != nil || digest != command.CommandDigest {
		t.Fatalf("ComputeCommandDigest = %q, %v", digest, err)
	}
	target := resourceTarget("server")
	if ComputeNativeRefHash(target.NativeRef) == "" || ComputeEvidenceSubjectHash(command, target, ObservationPresent) == "" {
		t.Fatal("evidence digest helper returned empty digest")
	}
	receipt := mustReceipt(t, command, Receipt{Status: StatusSucceeded, Phase: PhasePlanned}, nil)
	if digest, err := ComputeReceiptDigest(receipt); err != nil || digest != receipt.ReceiptDigest {
		t.Fatalf("ComputeReceiptDigest = %q, %v", digest, err)
	}
	receipt.AdapterID = "mutated"
	if err := receipt.ValidateFor(context.Background(), command, nil); !errors.Is(err, ErrInvalidReceipt) {
		t.Fatalf("ValidateFor error = %v, want ErrInvalidReceipt", err)
	}
}

func TestExecutorContractReturnsSingleUnchainedResult(t *testing.T) {
	var executor Executor = staticExecutor{}
	result := executor.Execute(context.Background(), ExecutionRequest{})
	if result.Status != StatusDenied {
		t.Fatalf("status = %q, want %q", result.Status, StatusDenied)
	}
}

type recordingVerifier struct {
	calls int
	err   error
}

func (v *recordingVerifier) VerifyEvidence(_ context.Context, _ Command, _ ResourceTarget, _ Evidence) error {
	v.calls++
	return v.err
}

type staticExecutor struct{}

func (staticExecutor) Execute(context.Context, ExecutionRequest) ExecutionResult {
	return ExecutionResult{Status: StatusDenied}
}

func validCommand(operation Operation) Command {
	command := Command{
		Operation:              operation,
		TenantID:               "tenant-1",
		LeaseID:                "lease-1",
		LeaseRevision:          11,
		RuntimeServerID:        "server-1",
		ResourceGenerationID:   contractResourceGenerationID,
		IdempotencyKey:         "request-1",
		ProviderID:             "centron",
		AdapterID:              "managed-compute",
		CustodyRef:             "custody://tenant-1/provider-access",
		CustodyHash:            testDigest("custody"),
		ConnectionRef:          "provider-connection://tenant-1/managed-compute",
		ConnectionHash:         testDigest("connection"),
		CapabilitySnapshotHash: testDigest("capabilities"),
		ExecutionProfileHash:   testDigest("profile"),
		ResourceGraphHash:      testDigest("graph"),
		LedgerRevision:         7,
		RequestedAt:            contractNow,
	}
	if operation == OperationPlan || operation == OperationProvision || operation == OperationReconcile {
		command.DesiredSpecRef = "desired-spec://techstack/leases/lease-1"
		command.DesiredSpecHash = testDigest("spec")
	}
	if operation == OperationObserve || operation == OperationReconcile || operation == OperationDecommission {
		command.Targets = []ResourceTarget{resourceTarget("server")}
	}
	return command
}

func resourceTarget(bindingID string) ResourceTarget {
	return ResourceTarget{
		BindingID:     bindingID,
		Kind:          "compute",
		NativeRef:     "native-" + bindingID,
		OwnershipHash: testDigest("owner-" + bindingID),
		Disposition:   DispositionDelete,
	}
}

func targetsForTest(count int) []ResourceTarget {
	targets := make([]ResourceTarget, 0, count)
	for i := 0; i < count; i++ {
		targets = append(targets, resourceTarget(fmt.Sprintf("target-%03d", i)))
	}
	return targets
}

func resourceBinding(target ResourceTarget, observation ObservationState, cleanup CleanupState) ResourceBinding {
	return ResourceBinding{
		BindingID:       target.BindingID,
		Kind:            target.Kind,
		NativeRef:       target.NativeRef,
		ParentBindingID: target.ParentBindingID,
		OwnershipHash:   target.OwnershipHash,
		Disposition:     target.Disposition,
		Observation:     observation,
		Cleanup:         cleanup,
	}
}

func absentBinding(command Command, target ResourceTarget, suffix string) ResourceBinding {
	resource := resourceBinding(target, ObservationAbsent, CleanupComplete)
	resource.Evidence = []Evidence{absenceEvidence(command, target, suffix)}
	return resource
}

func absenceEvidence(command Command, target ResourceTarget, suffix string) Evidence {
	return observationEvidence(command, target, ObservationAbsent, suffix)
}

func observationEvidence(command Command, target ResourceTarget, observation ObservationState, suffix string) Evidence {
	return Evidence{
		Ref:                    "provider-evidence://operations/" + suffix,
		Digest:                 testDigest("evidence-" + suffix),
		Source:                 EvidenceSourceProviderAPI,
		OperationID:            command.OperationID,
		LeaseRevision:          command.LeaseRevision,
		RuntimeServerID:        command.RuntimeServerID,
		ProviderID:             command.ProviderID,
		CapabilitySnapshotHash: command.CapabilitySnapshotHash,
		BindingID:              target.BindingID,
		NativeRefHash:          ComputeNativeRefHash(target.NativeRef),
		ConnectionHash:         command.ConnectionHash,
		ExecutionProfileHash:   command.ExecutionProfileHash,
		ResourceGraphHash:      command.ResourceGraphHash,
		SubjectHash:            ComputeEvidenceSubjectHash(command, target, observation),
		Observation:            observation,
		Definitive:             true,
		AttestationRef:         "provider-attestation://operations/" + suffix,
		AttestationDigest:      testDigest("attestation-" + suffix),
		CollectedAt:            command.RequestedAt.Add(30 * time.Second),
	}
}

func mustCommand(t *testing.T, command Command) Command {
	t.Helper()
	sealed, err := SealCommand(command)
	if err != nil {
		t.Fatalf("SealCommand: %v", err)
	}
	return sealed
}

func mustReceipt(t *testing.T, command Command, receipt Receipt, verifier EvidenceVerifier) Receipt {
	t.Helper()
	sealed, err := sealReceipt(t, command, receipt, verifier)
	if err != nil {
		t.Fatalf("receipt envelope: %v", err)
	}
	return sealed
}

func sealReceipt(t *testing.T, command Command, receipt Receipt, verifier EvidenceVerifier) (Receipt, error) {
	t.Helper()
	issuedAt := command.RequestedAt.Add(time.Minute)
	if !receipt.IssuedAt.IsZero() {
		issuedAt = receipt.IssuedAt
	}
	return sealReceiptAt(t, command, receipt, verifier, issuedAt)
}

func sealReceiptAt(t *testing.T, command Command, receipt Receipt, verifier EvidenceVerifier, issuedAt time.Time) (Receipt, error) {
	t.Helper()
	receipt = receiptForCommand(command, receipt)
	receipt.IssuedAt = issuedAt
	return sealReceiptEnvelope(context.Background(), command, receipt, verifier)
}

func receiptForCommand(command Command, receipt Receipt) Receipt {
	if receipt.APIVersion == "" {
		receipt.APIVersion = APIVersion
	}
	receipt.OperationID = command.OperationID
	receipt.Operation = command.Operation
	receipt.TenantID = command.TenantID
	receipt.LeaseID = command.LeaseID
	receipt.LeaseRevision = command.LeaseRevision
	receipt.RuntimeServerID = command.RuntimeServerID
	receipt.ResourceGenerationID = command.ResourceGenerationID
	receipt.IdempotencyKey = command.IdempotencyKey
	receipt.ProviderID = command.ProviderID
	receipt.AdapterID = command.AdapterID
	receipt.CapabilitySnapshotHash = command.CapabilitySnapshotHash
	receipt.CommandDigest = command.CommandDigest
	receipt.DesiredSpecHash = command.DesiredSpecHash
	if receipt.Sequence == 0 {
		// Most unit tests exercise a standalone post-request receipt. The
		// coordinator-owned initial pending/requested entry is tested explicitly.
		receipt.Sequence = 2
		receipt.PreviousReceiptDigest = testDigest("test-initial-receipt")
	}
	if receipt.IssuedAt.IsZero() {
		receipt.IssuedAt = command.RequestedAt.Add(time.Minute)
	}
	if receipt.PhaseEnteredAt.IsZero() {
		receipt.PhaseEnteredAt = receipt.IssuedAt
	}
	return receipt
}

func resealReceipt(t *testing.T, receipt Receipt) Receipt {
	t.Helper()
	receipt.ReceiptDigest = ""
	digest, err := ComputeReceiptDigest(receipt)
	if err != nil {
		t.Fatalf("ComputeReceiptDigest: %v", err)
	}
	receipt.ReceiptDigest = digest
	return receipt
}

func cloneResources(resources []ResourceBinding) []ResourceBinding {
	cloned := append([]ResourceBinding(nil), resources...)
	for i := range cloned {
		cloned[i].Evidence = append([]Evidence(nil), resources[i].Evidence...)
	}
	return cloned
}

func testDigest(value string) string {
	return sha256Digest([]byte(value))
}
