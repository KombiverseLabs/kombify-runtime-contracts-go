package runtimelease

import (
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"testing"
	"time"
)

const (
	testSnapshotKeyID    = "lease-key-2026-07"
	testSnapshotIssuer   = "techstack-lease-authority"
	testSnapshotAudience = "runtime-enrollment"
)

func validTestLease(now time.Time) Lease {
	return Lease{
		ID:                   "lease-1",
		Revision:             7,
		TenantID:             "org-1",
		OwnerID:              "user-1",
		ServerID:             "server-1",
		ResourceGenerationID: testResourceGenerationID,
		DesiredState:         DesiredStateRunning,
		ValidFrom:            now.Add(-time.Minute),
		ValidUntil:           now.Add(time.Hour),
	}
}

func testSnapshotKeys(t *testing.T) (ed25519.PublicKey, ed25519.PrivateKey) {
	t.Helper()
	seed := sha256.Sum256([]byte(t.Name()))
	privateKey := ed25519.NewKeyFromSeed(seed[:])
	return privateKey.Public().(ed25519.PublicKey), privateKey
}

func testSnapshot(now time.Time) Snapshot {
	return Snapshot{
		Issuer: testSnapshotIssuer, Audience: testSnapshotAudience,
		Lease: validTestLease(now), IssuedAt: now, ExpiresAt: now.Add(4 * time.Minute),
	}
}

func testSnapshotBinding(lease Lease) SnapshotBinding {
	return SnapshotBinding{
		Issuer: testSnapshotIssuer, Audience: testSnapshotAudience,
		TenantID: lease.TenantID, OwnerID: lease.OwnerID, LeaseID: lease.ID,
		LeaseRevision: lease.Revision, ServerID: lease.ServerID,
		ResourceGenerationID: lease.ResourceGenerationID,
	}
}

func testSnapshotKeySet(t *testing.T, publicKey ed25519.PublicKey, now time.Time) *SnapshotKeySet {
	t.Helper()
	set, err := NewSnapshotKeySet([]SnapshotVerificationKey{{
		KeyID: testSnapshotKeyID, Issuer: testSnapshotIssuer, PublicKey: publicKey,
		NotBefore: now.Add(-time.Hour), NotAfter: now.Add(time.Hour),
	}})
	if err != nil {
		t.Fatalf("NewSnapshotKeySet: %v", err)
	}
	return set
}

func TestSignAndVerifySnapshotRoundTrip(t *testing.T) {
	now := time.Date(2026, 5, 6, 12, 0, 0, 0, time.UTC)
	publicKey, privateKey := testSnapshotKeys(t)
	snapshot := testSnapshot(now)

	signed, err := SignSnapshot(snapshot, testSnapshotKeyID, privateKey)
	if err != nil {
		t.Fatalf("SignSnapshot: %v", err)
	}
	if signed.Version != SnapshotVersion || signed.Algorithm != SnapshotAlgorithm ||
		signed.KeyID != testSnapshotKeyID || signed.Signature == "" {
		t.Fatalf("signed snapshot missing versioned identity: %+v", signed)
	}

	lease, err := VerifySnapshot(signed, testSnapshotKeySet(t, publicKey, now), testSnapshotBinding(snapshot.Lease), now.Add(time.Minute))
	if err != nil {
		t.Fatalf("VerifySnapshot: %v", err)
	}
	if lease.ID != snapshot.Lease.ID || lease.Revision != snapshot.Lease.Revision || lease.ServerID != "server-1" ||
		lease.ResourceGenerationID != testResourceGenerationID {
		t.Fatalf("unexpected lease from snapshot: %+v", lease)
	}
}

func TestSnapshotVerificationKeyCannotMintSnapshots(t *testing.T) {
	_, privateKey := testSnapshotKeys(t)
	publicKey := privateKey.Public().(ed25519.PublicKey)
	now := time.Date(2026, 7, 21, 12, 0, 0, 0, time.UTC)
	_, err := SignSnapshot(testSnapshot(now), testSnapshotKeyID, ed25519.PrivateKey(publicKey))
	if !errors.Is(err, ErrSnapshotKey) {
		t.Fatalf("SignSnapshot with public key error = %v, want ErrSnapshotKey", err)
	}
}

func TestVerifySnapshotRejectsTypedNilKeyResolver(t *testing.T) {
	now := time.Date(2026, 5, 6, 12, 0, 0, 0, time.UTC)
	_, privateKey := testSnapshotKeys(t)
	signed, err := SignSnapshot(testSnapshot(now), testSnapshotKeyID, privateKey)
	if err != nil {
		t.Fatalf("SignSnapshot: %v", err)
	}
	var resolver *SnapshotKeySet
	if _, err := VerifySnapshot(signed, resolver, testSnapshotBinding(signed.Lease), now); !errors.Is(err, ErrSnapshotKey) {
		t.Fatalf("typed-nil resolver error = %v, want ErrSnapshotKey", err)
	}
}

func TestSnapshotKeySetCopiesKeysAndRejectsMutation(t *testing.T) {
	now := time.Date(2026, 7, 21, 12, 0, 0, 0, time.UTC)
	publicKey, privateKey := testSnapshotKeys(t)
	original := append(ed25519.PublicKey(nil), publicKey...)
	set := testSnapshotKeySet(t, publicKey, now)
	for index := range publicKey {
		publicKey[index] = 0
	}
	signed, err := SignSnapshot(testSnapshot(now), testSnapshotKeyID, privateKey)
	if err != nil {
		t.Fatalf("SignSnapshot: %v", err)
	}
	if _, err := VerifySnapshot(signed, set, testSnapshotBinding(signed.Lease), now); err != nil {
		t.Fatalf("stored key changed with caller slice: %v", err)
	}
	resolved, err := set.ResolveSnapshotVerificationKey(testSnapshotKeyID)
	if err != nil {
		t.Fatalf("ResolveSnapshotVerificationKey: %v", err)
	}
	resolved.PublicKey[0] ^= 0xff
	again, _ := set.ResolveSnapshotVerificationKey(testSnapshotKeyID)
	if string(again.PublicKey) != string(original) {
		t.Fatal("resolved key mutation changed immutable key set")
	}
}

func TestVerifySnapshotRejectsTamperingWrongAudienceAndWrongKey(t *testing.T) {
	now := time.Date(2026, 5, 6, 12, 0, 0, 0, time.UTC)
	publicKey, privateKey := testSnapshotKeys(t)
	signed, err := SignSnapshot(testSnapshot(now), testSnapshotKeyID, privateKey)
	if err != nil {
		t.Fatalf("SignSnapshot: %v", err)
	}
	binding := testSnapshotBinding(signed.Lease)

	tampered := signed
	tampered.Lease.DesiredState = DesiredStateStopped
	if _, err := VerifySnapshot(tampered, testSnapshotKeySet(t, publicKey, now), binding, now.Add(time.Minute)); !errors.Is(err, ErrSnapshotSignature) {
		t.Fatalf("tampered VerifySnapshot error = %v, want ErrSnapshotSignature", err)
	}

	wrongAudience := binding
	wrongAudience.Audience = "different-consumer"
	if _, err := VerifySnapshot(signed, testSnapshotKeySet(t, publicKey, now), wrongAudience, now.Add(time.Minute)); !errors.Is(err, ErrBindingMismatch) {
		t.Fatalf("wrong-audience VerifySnapshot error = %v, want ErrBindingMismatch", err)
	}

	otherSeed := sha256.Sum256([]byte("other-key"))
	otherPrivate := ed25519.NewKeyFromSeed(otherSeed[:])
	otherPublic := otherPrivate.Public().(ed25519.PublicKey)
	if _, err := VerifySnapshot(signed, testSnapshotKeySet(t, otherPublic, now), binding, now.Add(time.Minute)); !errors.Is(err, ErrSnapshotSignature) {
		t.Fatalf("wrong-key VerifySnapshot error = %v, want ErrSnapshotSignature", err)
	}
}

func TestSnapshotRejectsInvalidUTF8BeforeSigning(t *testing.T) {
	now := time.Date(2026, 7, 21, 12, 0, 0, 0, time.UTC)
	_, privateKey := testSnapshotKeys(t)
	snapshot := testSnapshot(now)
	snapshot.Lease.TenantID = string([]byte{0xff})
	if _, err := SignSnapshot(snapshot, testSnapshotKeyID, privateKey); !errors.Is(err, ErrSnapshotInvalid) {
		t.Fatalf("invalid UTF-8 SignSnapshot error = %v, want ErrSnapshotInvalid", err)
	}
}

func TestSignSnapshotRejectsInvalidProjectionTTLAndGrace(t *testing.T) {
	now := time.Date(2026, 7, 21, 12, 0, 0, 0, time.UTC)
	_, privateKey := testSnapshotKeys(t)
	snapshot := testSnapshot(now)
	snapshot.Lease.ResourceGenerationID = ResourceGenerationID("550E8400-E29B-41D4-A716-446655440000")
	if _, err := SignSnapshot(snapshot, testSnapshotKeyID, privateKey); !errors.Is(err, ErrSnapshotInvalid) {
		t.Fatalf("invalid lease SignSnapshot error = %v, want ErrSnapshotInvalid", err)
	}

	snapshot = testSnapshot(now)
	snapshot.ExpiresAt = now.Add(MaximumSnapshotTTL + time.Nanosecond)
	if _, err := SignSnapshot(snapshot, testSnapshotKeyID, privateKey); !errors.Is(err, ErrSnapshotInvalid) {
		t.Fatalf("oversized TTL SignSnapshot error = %v, want ErrSnapshotInvalid", err)
	}

	snapshot = testSnapshot(now)
	tooLate := snapshot.ExpiresAt.Add(MaximumSnapshotGrace + time.Nanosecond)
	snapshot.GraceUntil = &tooLate
	if _, err := SignSnapshot(snapshot, testSnapshotKeyID, privateKey); !errors.Is(err, ErrSnapshotInvalid) {
		t.Fatalf("oversized grace SignSnapshot error = %v, want ErrSnapshotInvalid", err)
	}
}

func TestVerifySnapshotRejectsFutureIssueAndMalformedSignedShape(t *testing.T) {
	now := time.Date(2026, 7, 21, 12, 0, 0, 0, time.UTC)
	publicKey, privateKey := testSnapshotKeys(t)
	snapshot := testSnapshot(now)
	snapshot.IssuedAt = now.Add(time.Minute)
	snapshot.ExpiresAt = now.Add(2 * time.Minute)
	signed, err := SignSnapshot(snapshot, testSnapshotKeyID, privateKey)
	if err != nil {
		t.Fatalf("SignSnapshot: %v", err)
	}
	if _, err := VerifySnapshot(signed, testSnapshotKeySet(t, publicKey, now), testSnapshotBinding(signed.Lease), now); !errors.Is(err, ErrSnapshotInvalid) {
		t.Fatalf("future-issued VerifySnapshot error = %v, want ErrSnapshotInvalid", err)
	}

	malformed := signed
	malformed.IssuedAt = now.Add(3 * time.Minute)
	malformed.ExpiresAt = now.Add(2 * time.Minute)
	malformed.Signature = ""
	payload, err := snapshotSigningPayload(malformed)
	if err != nil {
		t.Fatalf("snapshotSigningPayload: %v", err)
	}
	malformed.Signature = base64.RawURLEncoding.EncodeToString(ed25519.Sign(privateKey, payload))
	if _, err := VerifySnapshot(malformed, testSnapshotKeySet(t, publicKey, now), testSnapshotBinding(malformed.Lease), now); !errors.Is(err, ErrSnapshotInvalid) {
		t.Fatalf("malformed signed shape error = %v, want ErrSnapshotInvalid", err)
	}
}

func TestAuthoritySignedGraceIsBoundedAndCannotBeCallerExtended(t *testing.T) {
	now := time.Date(2026, 5, 6, 12, 0, 0, 0, time.UTC)
	publicKey, privateKey := testSnapshotKeys(t)
	snapshot := testSnapshot(now)
	snapshot.Lease.ValidFrom = now.Add(-20 * time.Minute)
	snapshot.IssuedAt = now.Add(-5 * time.Minute)
	snapshot.ExpiresAt = now.Add(-time.Minute)
	graceUntil := snapshot.ExpiresAt.Add(3 * time.Minute)
	snapshot.GraceUntil = &graceUntil
	signed, err := SignSnapshot(snapshot, testSnapshotKeyID, privateKey)
	if err != nil {
		t.Fatalf("SignSnapshot: %v", err)
	}
	set := testSnapshotKeySet(t, publicKey, now)
	binding := testSnapshotBinding(signed.Lease)
	outage := &AuthorityError{Operation: "validate", Kind: AuthorityErrorStatus, StatusCode: 503}
	if _, err := VerifySnapshot(signed, set, binding, now); !errors.Is(err, ErrSnapshotExpired) {
		t.Fatalf("normal VerifySnapshot error = %v, want ErrSnapshotExpired", err)
	}
	if _, err := VerifySnapshotFallback(signed, set, binding, now, outage); err != nil {
		t.Fatalf("VerifySnapshotFallback: %v", err)
	}
	if _, err := VerifySnapshotFallback(signed, set, binding, graceUntil, outage); !errors.Is(err, ErrSnapshotGrace) {
		t.Fatalf("past signed grace error = %v, want ErrSnapshotGrace", err)
	}
	authoritativeConflict := &AuthorityError{Operation: "validate", Kind: AuthorityErrorStatus, StatusCode: 409}
	if _, err := VerifySnapshotFallback(signed, set, binding, now, authoritativeConflict); !errors.Is(err, ErrSnapshotGrace) {
		t.Fatalf("authoritative rejection fallback error = %v, want ErrSnapshotGrace", err)
	}

	withoutGrace := snapshot
	withoutGrace.GraceUntil = nil
	withoutGrace, err = SignSnapshot(withoutGrace, testSnapshotKeyID, privateKey)
	if err != nil {
		t.Fatalf("SignSnapshot without grace: %v", err)
	}
	if _, err := VerifySnapshotFallback(withoutGrace, set, binding, now, outage); !errors.Is(err, ErrSnapshotGrace) {
		t.Fatalf("unsigned grace extension error = %v, want ErrSnapshotGrace", err)
	}
}
