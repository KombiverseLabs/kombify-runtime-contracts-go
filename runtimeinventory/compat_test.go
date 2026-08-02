package runtimeinventory

import (
	"reflect"
	"strings"
	"testing"
)

func TestOpenAPIJSONFieldCompatibility(t *testing.T) {
	for _, contract := range []struct {
		value any
		tags  map[string]string
	}{
		{Freshness{}, map[string]string{"State": "state", "AgeSeconds": "age_seconds,omitempty", "StaleAfterSeconds": "stale_after_seconds", "ReasonCode": "reason_code,omitempty"}},
		{Addresses{}, map[string]string{"PublicIPs": "public_ips", "PrivateIPs": "private_ips", "LocalIPs": "local_ips", "Domains": "domains"}},
		{ObservedState{}, map[string]string{"State": "state", "ReasonCode": "reason_code,omitempty", "ChangedAt": "changed_at,omitempty", "ObservedAt": "observed_at,omitempty"}},
		{LifecycleProjection{}, map[string]string{"State": "state", "ReasonCode": "reason_code,omitempty", "ChangedAt": "changed_at,omitempty"}},
		{DesiredProjection{}, map[string]string{"State": "state", "ReasonCode": "reason_code,omitempty", "ChangedAt": "changed_at,omitempty"}},
		{LeaseProjection{}, map[string]string{"ID": "id,omitempty", "State": "state"}},
		{CleanupProjection{}, map[string]string{"State": "state", "ProviderAbsenceVerified": "provider_absence_verified", "VerifiedAt": "verified_at,omitempty"}},
		{StackKit{}, map[string]string{"Name": "name,omitempty", "Version": "version,omitempty", "Variant": "variant,omitempty"}},
		{Platform{}, map[string]string{"OS": "os,omitempty", "OSVersion": "os_version,omitempty", "Arch": "arch,omitempty"}},
		{Server{}, map[string]string{"ID": "id", "StackID": "stack_id,omitempty", "Name": "name", "ObservedAt": "observed_at", "Freshness": "freshness", "InventoryRevision": "inventory_revision", "Addresses": "addresses", "Platform": "platform", "StackKit": "stackkit", "Provider": "provider,omitempty", "Lifecycle": "lifecycle", "Desired": "desired", "Lease": "lease", "Cleanup": "cleanup", "Connection": "connection", "Health": "health"}},
		{ServerList{}, map[string]string{"ObservedAt": "observed_at", "Freshness": "freshness", "InventoryRevision": "inventory_revision", "CollectionCursor": "collection_cursor", "NextCursor": "next_cursor,omitempty", "PageSize": "page_size", "Servers": "servers"}},
		{ServerHealth{}, map[string]string{"ServerID": "server_id", "ObservedAt": "observed_at", "Freshness": "freshness", "InventoryRevision": "inventory_revision", "Connection": "connection", "Health": "health"}},
		{ServiceLink{}, map[string]string{"URL": "url", "Mode": "mode"}},
		{Service{}, map[string]string{"ID": "id", "StackID": "stack_id", "ServerID": "server_id", "Key": "key", "Name": "name", "ObservedState": "observed_state", "ObservedAt": "observed_at", "Freshness": "freshness", "InventoryRevision": "inventory_revision", "Health": "health", "StackKit": "stackkit", "Links": "links", "Source": "source"}},
		{ServiceList{}, map[string]string{"ObservedAt": "observed_at", "Freshness": "freshness", "InventoryRevision": "inventory_revision", "CollectionCursor": "collection_cursor", "NextCursor": "next_cursor,omitempty", "PageSize": "page_size", "Services": "services"}},
		{Channel{}, map[string]string{"Type": "type", "Role": "role", "State": "state", "ObservedAt": "observed_at,omitempty"}},
		{AccessServiceLink{}, map[string]string{"ServiceID": "service_id", "ServiceName": "service_name", "URL": "url", "Mode": "mode", "Health": "health"}},
		{Availability{}, map[string]string{"State": "state", "ReasonCode": "reason_code,omitempty"}},
		{ServerAccessContext{}, map[string]string{"ServerID": "server_id", "ObservedAt": "observed_at", "Freshness": "freshness", "InventoryRevision": "inventory_revision", "Availability": "availability", "Connection": "connection", "Addresses": "addresses", "Channels": "channels", "ServiceLinks": "service_links"}},
	} {
		assertJSONTags(t, reflect.TypeOf(contract.value), contract.tags)
	}
}

func TestReadModelHasNoOpenMetadataOrSecretFields(t *testing.T) {
	for _, value := range []any{ServerList{}, ServerHealth{}, ServiceList{}, ServerAccessContext{}} {
		assertSecretFreeShape(t, reflect.TypeOf(value), map[reflect.Type]bool{})
	}
}

func assertJSONTags(t *testing.T, typ reflect.Type, expected map[string]string) {
	t.Helper()
	if typ.NumField() != len(expected) {
		t.Fatalf("%s has %d fields, want %d", typ.Name(), typ.NumField(), len(expected))
	}
	for index := 0; index < typ.NumField(); index++ {
		field := typ.Field(index)
		if got, ok := expected[field.Name]; !ok || field.Tag.Get("json") != got {
			t.Errorf("%s.%s json tag = %q, want %q", typ.Name(), field.Name, field.Tag.Get("json"), got)
		}
	}
}

func assertSecretFreeShape(t *testing.T, typ reflect.Type, seen map[reflect.Type]bool) {
	t.Helper()
	for typ.Kind() == reflect.Pointer {
		typ = typ.Elem()
	}
	if typ.Kind() == reflect.Slice || typ.Kind() == reflect.Array {
		if typ.Elem().Kind() == reflect.Uint8 {
			t.Fatalf("secret-free contract contains byte payload %s", typ)
		}
		assertSecretFreeShape(t, typ.Elem(), seen)
		return
	}
	if typ.PkgPath() == "time" || seen[typ] {
		return
	}
	seen[typ] = true
	if typ.Kind() == reflect.Map || typ.Kind() == reflect.Interface {
		t.Fatalf("secret-free contract contains open field type %s", typ)
	}
	if typ.Kind() != reflect.Struct || typ.PkgPath() != reflect.TypeOf(Server{}).PkgPath() {
		return
	}
	for index := 0; index < typ.NumField(); index++ {
		field := typ.Field(index)
		jsonName := strings.Split(field.Tag.Get("json"), ",")[0]
		for _, forbidden := range []string{"credential", "token", "secret", "metadata", "endpoint_ref", "route_id", "auth_hint", "ssh", "command", "logs"} {
			if strings.Contains(jsonName, forbidden) {
				t.Errorf("%s.%s exposes forbidden JSON field %q", typ.Name(), field.Name, jsonName)
			}
		}
		assertSecretFreeShape(t, field.Type, seen)
	}
}
