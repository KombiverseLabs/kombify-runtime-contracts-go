package stackaction

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestGeneratedStagingWireIsCanonicalStackAction(t *testing.T) {
	data, err := json.Marshal(Request{
		Action:  ActionStackKitRollout,
		StackID: "stack-1",
		RuntimeTarget: &RuntimeTarget{
			Host: "server.example",
			User: "root",
			AccessProfileRef: &ScopedReference{
				Ref:       "access-profile/stack-1/root",
				Version:   "v1",
				Scopes:    []string{"runtime:ssh"},
				ExpiresAt: time.Date(2099, 1, 1, 0, 0, 0, 0, time.UTC),
			},
		},
		PlatformNodes: []PlatformNode{{
			Name:     "worker-1",
			Role:     "worker",
			Platform: NodePlatformTarget{ServerID: "server-1"},
		}},
	})
	if err != nil {
		t.Fatalf("marshal generated StackAction request: %v", err)
	}
	wire := string(data)
	if !strings.Contains(wire, `"server_id"`) || strings.Contains(wire, `"serverId"`) {
		t.Fatalf("non-canonical generated wire: %s", wire)
	}
	if IsStackKitsAction(Action("simulate_update")) || IsStackKitsAction(Action("kit_upgrade")) {
		t.Fatal("generated staging output admits a non-executable action")
	}
	for _, field := range []string{
		"access_key_id", "agent_token", "channel_bootstrap", "client_private_key",
		"key_path", "key_pem", "komodo_onboarding_key", "password", "private_key",
		"repo_password", "secret_access_key", "secrets", "token",
	} {
		if strings.Contains(wire, `"`+field+`":`) {
			t.Fatalf("generated staging wire exposed forbidden field %q: %s", field, wire)
		}
	}
}
