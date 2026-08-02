package runtimeinventory

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"testing"
)

const (
	runtimeInventoryManifestVersion  = "techstack.runtimeinventory-generation-manifest/v1"
	runtimeInventoryGeneratorVersion = "runtimeinventorygen/v1"
	runtimeInventoryWireVersion      = "runtimeinventory/v1"
	runtimeInventorySourceRepository = "github.com/KombiverseLabs/kombify-Techstack"
	runtimeInventorySourcePath       = "contracts/runtimeinventory/v1/schema.json"
)

type runtimeInventoryManifest struct {
	SchemaVersion    string `json:"schemaVersion"`
	GeneratorVersion string `json:"generatorVersion"`
	WireVersion      string `json:"wireVersion"`
	Source           struct {
		Repository string `json:"repository"`
		Path       string `json:"path"`
		SHA256     string `json:"sha256"`
	} `json:"source"`
	Outputs map[string]string `json:"outputs"`
}

func TestGeneratedBundleHasExactTechstackLineage(t *testing.T) {
	manifestPayload, err := os.ReadFile("manifest.json")
	if err != nil {
		t.Fatal(err)
	}
	var manifest runtimeInventoryManifest
	decoder := json.NewDecoder(bytes.NewReader(manifestPayload))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&manifest); err != nil {
		t.Fatalf("decode generation manifest: %v", err)
	}
	if manifest.SchemaVersion != runtimeInventoryManifestVersion ||
		manifest.GeneratorVersion != runtimeInventoryGeneratorVersion ||
		manifest.WireVersion != runtimeInventoryWireVersion {
		t.Fatalf("unexpected generation identity: %#v", manifest)
	}
	if manifest.Source.Repository != runtimeInventorySourceRepository ||
		manifest.Source.Path != runtimeInventorySourcePath {
		t.Fatalf("unexpected source authority: %#v", manifest.Source)
	}
	assertFileDigest(t, "schema.json", manifest.Source.SHA256)
	if len(manifest.Outputs) != 1 {
		t.Fatalf("generation manifest has %d outputs, want exactly types.go", len(manifest.Outputs))
	}
	typesDigest, ok := manifest.Outputs["types.go"]
	if !ok {
		t.Fatal("generation manifest does not bind types.go")
	}
	assertFileDigest(t, "types.go", typesDigest)
}

func assertFileDigest(t *testing.T, path, expected string) {
	t.Helper()
	payload, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(payload)
	if got := hex.EncodeToString(sum[:]); got != expected {
		t.Fatalf("%s SHA-256 = %s, want %s", path, got, expected)
	}
}
