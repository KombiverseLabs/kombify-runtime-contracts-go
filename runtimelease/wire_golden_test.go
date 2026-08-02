package runtimelease

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestLeaseWireGolden(t *testing.T) {
	validFrom := time.Date(2026, 7, 21, 12, 0, 0, 0, time.UTC)
	validUntil := validFrom.Add(24 * time.Hour)
	payload, err := json.Marshal(Lease{
		ID:                   "lease-1",
		Revision:             7,
		TenantID:             "org-1",
		OwnerID:              "user-1",
		ServerID:             "server-1",
		ResourceGenerationID: testResourceGenerationID,
		DesiredState:         DesiredStateRunning,
		ValidFrom:            validFrom,
		ValidUntil:           validUntil,
	})
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	want := `{"id":"lease-1","revision":7,"tenant_id":"org-1","owner_id":"user-1","server_id":"server-1","resource_generation_id":"550e8400-e29b-41d4-a716-446655440000","desired_state":"running","valid_from":"2026-07-21T12:00:00Z","valid_until":"2026-07-22T12:00:00Z"}`
	if string(payload) != want {
		t.Fatalf("lease wire format = %s, want %s", payload, want)
	}
}

func TestValidationResultReasonWireGolden(t *testing.T) {
	payload, err := json.Marshal(ValidationResult{Valid: false, Reason: ReasonBindingMismatch})
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	want := `{"valid":false,"reason":"binding_mismatch"}`
	if string(payload) != want {
		t.Fatalf("validation result wire format = %s, want %s", payload, want)
	}
}

func TestValidationCheckWireGolden(t *testing.T) {
	payload, err := json.Marshal(ValidationCheck{
		TenantID:             "org-1",
		OwnerID:              "user-1",
		LeaseID:              "lease-1",
		LeaseRevision:        7,
		ServerID:             "server-1",
		ResourceGenerationID: testResourceGenerationID,
	})
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	want := `{"tenant_id":"org-1","owner_id":"user-1","lease_id":"lease-1","lease_revision":7,"server_id":"server-1","resource_generation_id":"550e8400-e29b-41d4-a716-446655440000"}`
	if string(payload) != want {
		t.Fatalf("validation check wire format = %s, want %s", payload, want)
	}
}

func TestEnrollmentRequestWireGolden(t *testing.T) {
	payload, err := json.Marshal(EnrollmentRequest{
		TenantID:             "org-1",
		OwnerID:              "user-1",
		LeaseID:              "lease-1",
		LeaseRevision:        7,
		ServerID:             "server-1",
		ResourceGenerationID: testResourceGenerationID,
		IdempotencyKey:       "enroll:lease-1:generation-1",
	})
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	want := `{"tenant_id":"org-1","owner_id":"user-1","lease_id":"lease-1","lease_revision":7,"server_id":"server-1","resource_generation_id":"550e8400-e29b-41d4-a716-446655440000","idempotency_key":"enroll:lease-1:generation-1"}`
	if string(payload) != want {
		t.Fatalf("enrollment request wire format = %s, want %s", payload, want)
	}
}

func TestEnrollmentResultWireGolden(t *testing.T) {
	payload, err := json.Marshal(EnrollmentResult{
		TenantID:             "org-1",
		OwnerID:              "user-1",
		LeaseID:              "lease-1",
		LeaseRevision:        7,
		ServerID:             "server-1",
		ResourceGenerationID: testResourceGenerationID,
	})
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	want := `{"tenant_id":"org-1","owner_id":"user-1","lease_id":"lease-1","lease_revision":7,"server_id":"server-1","resource_generation_id":"550e8400-e29b-41d4-a716-446655440000"}`
	if string(payload) != want {
		t.Fatalf("enrollment result wire format = %s, want %s", payload, want)
	}
}

func TestPublicLeaseAndEnrollmentWireContainsNoProviderOrSecretState(t *testing.T) {
	now := time.Date(2026, 7, 21, 12, 0, 0, 0, time.UTC)
	payloads := []any{
		validTestLease(now),
		EnrollmentRequest{
			TenantID:             "org-1",
			OwnerID:              "user-1",
			LeaseID:              "lease-1",
			LeaseRevision:        7,
			ServerID:             "server-1",
			ResourceGenerationID: testResourceGenerationID,
			IdempotencyKey:       "enroll:lease-1:generation-1",
		},
		EnrollmentResult{
			TenantID:             "org-1",
			OwnerID:              "user-1",
			LeaseID:              "lease-1",
			LeaseRevision:        7,
			ServerID:             "server-1",
			ResourceGenerationID: testResourceGenerationID,
		},
	}
	banned := []string{
		`"provider_id"`, `"adapter_id"`, `"simulation_id"`, `"vm_id"`,
		`"engine_vm_id"`, `"native_ref"`, `"observed_state"`,
		`"provisioning_spec"`, `"credential`, `"secret`, `"access_`,
	}
	for _, value := range payloads {
		payload, err := json.Marshal(value)
		if err != nil {
			t.Fatalf("Marshal(%T): %v", value, err)
		}
		for _, field := range banned {
			if strings.Contains(string(payload), field) {
				t.Fatalf("%T wire contains forbidden state %q: %s", value, field, payload)
			}
		}
	}
}
