package runtimeinventory

import (
	"reflect"
	"strings"
	"testing"
)

func TestReadModelHasNoOpenMetadataOrSecretFields(t *testing.T) {
	for _, value := range []any{ServerList{}, ServerHealth{}, ServiceList{}, ServerAccessContext{}} {
		assertSecretFreeShape(t, reflect.TypeOf(value), map[reflect.Type]bool{})
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
