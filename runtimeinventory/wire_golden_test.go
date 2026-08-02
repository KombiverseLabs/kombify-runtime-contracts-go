package runtimeinventory

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"
	"time"
)

func TestWireGolden(t *testing.T) {
	observedAt := time.Date(2026, 7, 21, 10, 11, 12, 123456789, time.UTC)
	age := int64(12)
	freshness := Freshness{State: "fresh", AgeSeconds: &age, StaleAfterSeconds: 90}
	addresses := Addresses{
		PublicIPs: []string{"203.0.113.10"}, PrivateIPs: []string{"10.0.0.4"},
		LocalIPs: []string{"192.168.1.4"}, Domains: []string{"runtime.example.test"},
	}
	connection := ObservedState{State: "connected", ReasonCode: "guard_connected", ChangedAt: &observedAt, ObservedAt: &observedAt}
	health := ObservedState{State: "healthy", ReasonCode: "checks_passing", ChangedAt: &observedAt, ObservedAt: &observedAt}
	server := Server{
		ID: "server-1", StackID: "stack-1", Name: "Runtime One", ObservedAt: &observedAt,
		Freshness: freshness, InventoryRevision: 7, Addresses: addresses,
		Platform: Platform{OS: "ubuntu", OSVersion: "24.04", Arch: "amd64"},
		StackKit: StackKit{Name: "cloud-kit", Version: "1.2.3", Variant: "main"}, Provider: "ionos",
		Lifecycle: LifecycleProjection{State: "active", ReasonCode: "enrolled", ChangedAt: &observedAt},
		Desired:   DesiredProjection{State: "running", ReasonCode: "operator_requested", ChangedAt: &observedAt}, Lease: LeaseProjection{ID: "lease-1", State: "active"},
		Cleanup:    CleanupProjection{State: "absent", ProviderAbsenceVerified: true, VerifiedAt: &observedAt},
		Connection: connection, Health: health,
	}
	service := Service{
		ID: "service-1", StackID: "stack-1", ServerID: "server-1", Key: "coolify", Name: "Coolify",
		ObservedState: "running", ObservedAt: &observedAt, Freshness: freshness, InventoryRevision: 7,
		Health: health, StackKit: StackKit{Version: "1.2.3"},
		Links: []ServiceLink{{URL: "https://coolify.example.test", Mode: "relay"}}, Source: "guard-inventory",
	}
	fixture := struct {
		ServerPage    ServerList          `json:"server_page"`
		ServicePage   ServiceList         `json:"service_page"`
		Health        ServerHealth        `json:"health"`
		AccessContext ServerAccessContext `json:"access_context"`
	}{
		ServerPage:  ServerList{ObservedAt: observedAt, Freshness: freshness, InventoryRevision: 7, CollectionCursor: "servers-cursor", NextCursor: "servers-next", PageSize: 1, Servers: []Server{server}},
		ServicePage: ServiceList{ObservedAt: observedAt, Freshness: freshness, InventoryRevision: 7, CollectionCursor: "services-cursor", NextCursor: "services-next", PageSize: 1, Services: []Service{service}},
		Health:      ServerHealth{ServerID: "server-1", ObservedAt: &observedAt, Freshness: freshness, InventoryRevision: 7, Connection: connection, Health: health},
		AccessContext: ServerAccessContext{
			ServerID: "server-1", ObservedAt: &observedAt, Freshness: freshness, InventoryRevision: 7,
			Availability: Availability{State: "available"}, Connection: connection, Addresses: addresses,
			Channels:     []Channel{{Type: "guard-api", Role: "primary", State: "connected", ObservedAt: &observedAt}},
			ServiceLinks: []AccessServiceLink{{ServiceID: "service-1", ServiceName: "Coolify", URL: "https://coolify.example.test", Mode: "relay", Health: "healthy"}},
		},
	}

	got, err := json.MarshalIndent(fixture, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	got = append(got, '\n')
	want, err := os.ReadFile("testdata/runtimeinventory.golden.json")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("runtimeinventory wire format changed\n--- got ---\n%s\n--- want ---\n%s", got, want)
	}
}
