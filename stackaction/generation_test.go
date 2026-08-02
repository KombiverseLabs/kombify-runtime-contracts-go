package stackaction

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"testing"
)

const generationManifestSchemaV1 = "stackkit.stackaction-generation-manifest/v1"

type generationManifest struct {
	SchemaVersion    string `json:"schemaVersion"`
	GeneratorVersion string `json:"generatorVersion"`
	WireVersion      string `json:"wireVersion"`
	Source           struct {
		Path   string `json:"path"`
		SHA256 string `json:"sha256"`
	} `json:"source"`
	Outputs map[string]string `json:"outputs"`
}

func TestGeneratedBundleManifest(t *testing.T) {
	t.Parallel()

	payload, err := os.ReadFile("manifest.json")
	if err != nil {
		t.Fatal(err)
	}
	var manifest generationManifest
	if err := json.Unmarshal(payload, &manifest); err != nil {
		t.Fatalf("decode generation manifest: %v", err)
	}
	if manifest.SchemaVersion != generationManifestSchemaV1 {
		t.Fatalf("schemaVersion = %q", manifest.SchemaVersion)
	}
	if manifest.GeneratorVersion != GeneratorVersion {
		t.Fatalf("generatorVersion = %q, generated Go = %q", manifest.GeneratorVersion, GeneratorVersion)
	}
	if manifest.WireVersion != WireVersion {
		t.Fatalf("wireVersion = %q, generated Go = %q", manifest.WireVersion, WireVersion)
	}
	if manifest.Source.Path != "base/stack_action.cue" || manifest.Source.SHA256 != SourceSHA256 {
		t.Fatalf("source = %+v, generated source hash = %q", manifest.Source, SourceSHA256)
	}

	wantOutputs := []string{"contract.ir.json", "openapi.yaml", "stackaction_gen.go"}
	if len(manifest.Outputs) != len(wantOutputs) {
		t.Fatalf("outputs = %v, want exactly %v", manifest.Outputs, wantOutputs)
	}
	for _, name := range wantOutputs {
		want := manifest.Outputs[name]
		if want == "" {
			t.Fatalf("manifest missing output digest for %s", name)
		}
		got, err := fileSHA256(name)
		if err != nil {
			t.Fatal(err)
		}
		if got != want {
			t.Fatalf("%s SHA-256 = %s, want %s", name, got, want)
		}
	}
}

func fileSHA256(path string) (string, error) {
	payload, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read %s: %w", path, err)
	}
	digest := sha256.Sum256(payload)
	return hex.EncodeToString(digest[:]), nil
}
