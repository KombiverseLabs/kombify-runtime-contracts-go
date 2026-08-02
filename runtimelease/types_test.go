package runtimelease

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"
)

const testResourceGenerationID ResourceGenerationID = "550e8400-e29b-41d4-a716-446655440000"

func TestLeaseValidateAcceptsAuthorityOwnedProjection(t *testing.T) {
	now := time.Date(2026, 7, 21, 12, 0, 0, 0, time.UTC)
	lease := validTestLease(now)

	if err := lease.Validate(now); err != nil {
		t.Fatalf("Validate returned error: %v", err)
	}
}

func TestLeaseValidateAcceptsCanonicalAbsentDesiredState(t *testing.T) {
	now := time.Date(2026, 7, 21, 12, 0, 0, 0, time.UTC)
	lease := validTestLease(now)
	lease.DesiredState = DesiredStateAbsent

	if err := lease.Validate(now); err != nil {
		t.Fatalf("Validate absent lease returned error: %v", err)
	}
}

func TestLeaseValidateRejectsUnknownDesiredState(t *testing.T) {
	now := time.Date(2026, 7, 21, 12, 0, 0, 0, time.UTC)
	lease := validTestLease(now)
	lease.DesiredState = DesiredState("archived")

	if err := lease.Validate(now); err != ErrInvalidLease {
		t.Fatalf("Validate unknown desired state error = %v, want ErrInvalidLease", err)
	}
}

func TestLeaseValidateRejectsExpiredLease(t *testing.T) {
	now := time.Date(2026, 7, 21, 12, 0, 0, 0, time.UTC)
	lease := validTestLease(now)
	lease.ValidFrom = now.Add(-2 * time.Hour)
	lease.ValidUntil = now.Add(-time.Minute)

	if err := lease.Validate(now); err != ErrLeaseExpired {
		t.Fatalf("Validate error = %v, want ErrLeaseExpired", err)
	}
}

func TestLeaseValidateRejectsEffectiveCancellation(t *testing.T) {
	now := time.Date(2026, 7, 21, 12, 0, 0, 0, time.UTC)
	cancelledAt := now.Add(-time.Second)
	lease := validTestLease(now)
	lease.CancelledAt = &cancelledAt

	if err := lease.Validate(now); err != ErrLeaseCancelled {
		t.Fatalf("Validate error = %v, want ErrLeaseCancelled", err)
	}
}

func TestLeaseValidateRequiresRevisionAndExactRuntimeBinding(t *testing.T) {
	now := time.Date(2026, 7, 21, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name   string
		mutate func(*Lease)
	}{
		{"revision", func(l *Lease) { l.Revision = 0 }},
		{"revision beyond exact wire range", func(l *Lease) { l.Revision = MaximumWireRevision + 1 }},
		{"tenant", func(l *Lease) { l.TenantID = "" }},
		{"owner", func(l *Lease) { l.OwnerID = "" }},
		{"server", func(l *Lease) { l.ServerID = "" }},
		{"generation missing", func(l *Lease) { l.ResourceGenerationID = "" }},
		{"generation nil UUID", func(l *Lease) { l.ResourceGenerationID = "00000000-0000-0000-0000-000000000000" }},
		{"generation uppercase", func(l *Lease) {
			l.ResourceGenerationID = ResourceGenerationID(strings.ToUpper(string(testResourceGenerationID)))
		}},
		{"generation noncanonical", func(l *Lease) { l.ResourceGenerationID = "550e8400e29b41d4a716446655440000" }},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			lease := validTestLease(now)
			test.mutate(&lease)
			if err := lease.Validate(now); err != ErrInvalidLease {
				t.Fatalf("Validate error = %v, want ErrInvalidLease", err)
			}
		})
	}
}

func TestLeaseActiveAtRequiresValidityWindow(t *testing.T) {
	now := time.Date(2026, 7, 21, 12, 0, 0, 0, time.UTC)
	lease := validTestLease(now)
	lease.ValidFrom = now.Add(time.Minute)
	lease.ValidUntil = now.Add(time.Hour)
	if err := lease.Validate(now); !errors.Is(err, ErrLeaseNotYetValid) {
		t.Fatalf("Validate before valid_from error = %v, want ErrLeaseNotYetValid", err)
	}
	if lease.ActiveAt(now) {
		t.Fatal("lease must not be active before valid_from")
	}
	if !lease.ActiveAt(now.Add(2 * time.Minute)) {
		t.Fatal("lease should be active inside its validity window")
	}
}

func TestResolveQueryRequiresTenantAndLease(t *testing.T) {
	if err := (ResolveQuery{TenantID: "org-1", OwnerID: "user-1", LeaseID: "lease-1"}).Validate(); err != nil {
		t.Fatalf("valid query: %v", err)
	}
	if err := (ResolveQuery{OwnerID: "user-1", LeaseID: "lease-1"}).Validate(); err != ErrInvalidLease {
		t.Fatalf("missing tenant error = %v, want ErrInvalidLease", err)
	}
	if err := (ResolveQuery{TenantID: "org-1", OwnerID: "user-1"}).Validate(); err != ErrInvalidLease {
		t.Fatalf("missing lease error = %v, want ErrInvalidLease", err)
	}
	if err := (ResolveQuery{TenantID: "org-1", LeaseID: "lease-1"}).Validate(); err != ErrInvalidLease {
		t.Fatalf("missing owner error = %v, want ErrInvalidLease", err)
	}
}

func TestValidationReasonForLeaseErrors(t *testing.T) {
	tests := []struct {
		err  error
		want string
	}{
		{ErrLeaseCancelled, ReasonLeaseCancelled},
		{ErrLeaseNotYetValid, ReasonLeaseNotYetValid},
		{ErrLeaseExpired, ReasonLeaseExpired},
		{ErrBindingMismatch, ReasonBindingMismatch},
		{ErrRevisionMismatch, ReasonRevisionMismatch},
		{ErrInvalidLease, ReasonInvalidLease},
	}
	for _, test := range tests {
		if got := ValidationReasonForError(test.err); got != test.want {
			t.Fatalf("ValidationReasonForError(%v) = %q, want %q", test.err, got, test.want)
		}
	}
}

func TestValidationCheckRequiresExactRevisionAndGeneration(t *testing.T) {
	valid := ValidationCheck{
		TenantID:             "org-1",
		OwnerID:              "user-1",
		LeaseID:              "lease-1",
		LeaseRevision:        7,
		ServerID:             "server-1",
		ResourceGenerationID: testResourceGenerationID,
	}
	if err := valid.Validate(); err != nil {
		t.Fatalf("valid check: %v", err)
	}

	invalidRevision := valid
	invalidRevision.LeaseRevision = 0
	if err := invalidRevision.Validate(); err != ErrInvalidLease {
		t.Fatalf("zero revision error = %v, want ErrInvalidLease", err)
	}
	invalidRevision.LeaseRevision = MaximumWireRevision + 1
	if err := invalidRevision.Validate(); err != ErrInvalidLease {
		t.Fatalf("oversized revision error = %v, want ErrInvalidLease", err)
	}

	invalidGeneration := valid
	invalidGeneration.ResourceGenerationID = ResourceGenerationID(strings.ToUpper(string(testResourceGenerationID)))
	if err := invalidGeneration.Validate(); err != ErrInvalidLease {
		t.Fatalf("uppercase generation error = %v, want ErrInvalidLease", err)
	}

}

func TestEnrollmentContractsRequireExactLeaseAndRuntimeBinding(t *testing.T) {
	request := EnrollmentRequest{
		TenantID:             "org-1",
		OwnerID:              "user-1",
		LeaseID:              "lease-1",
		LeaseRevision:        7,
		ServerID:             "server-1",
		ResourceGenerationID: testResourceGenerationID,
		IdempotencyKey:       "enroll:lease-1:generation-1",
	}
	if err := request.Validate(); err != nil {
		t.Fatalf("valid enrollment request: %v", err)
	}

	result := EnrollmentResult{
		TenantID:             request.TenantID,
		OwnerID:              request.OwnerID,
		LeaseID:              request.LeaseID,
		LeaseRevision:        request.LeaseRevision,
		ServerID:             request.ServerID,
		ResourceGenerationID: request.ResourceGenerationID,
	}
	if err := result.Validate(); err != nil {
		t.Fatalf("valid enrollment result: %v", err)
	}

	request.IdempotencyKey = ""
	if err := request.Validate(); err != ErrInvalidEnrollment {
		t.Fatalf("missing idempotency key error = %v, want ErrInvalidEnrollment", err)
	}
	request.IdempotencyKey = "enroll:lease-1:generation-1"
	request.LeaseRevision = MaximumWireRevision + 1
	if err := request.Validate(); err != ErrInvalidEnrollment {
		t.Fatalf("oversized request revision error = %v, want ErrInvalidEnrollment", err)
	}
	result.ResourceGenerationID = "not-a-uuid"
	if err := result.Validate(); err != ErrInvalidEnrollment {
		t.Fatalf("invalid result generation error = %v, want ErrInvalidEnrollment", err)
	}
}

func TestValidationCheckWireFormatIsProviderNeutral(t *testing.T) {
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
		t.Fatalf("wire format = %s, want %s", payload, want)
	}
}
