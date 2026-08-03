package contracts_test

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"

	providerexecutor "github.com/KombiverseLabs/kombify-runtime-contracts-go/providerexecutor/v1beta1"
	"github.com/KombiverseLabs/kombify-runtime-contracts-go/runtimeinventory"
	"github.com/KombiverseLabs/kombify-runtime-contracts-go/runtimelease"
	"github.com/KombiverseLabs/kombify-runtime-contracts-go/stackaction"
)

const corpusManifestSchemaV1 = "kombify.runtime-contract-fuzz-corpus/v1"

type corpusManifest struct {
	SchemaVersion string             `json:"schemaVersion"`
	Fixtures      []corpusFixture    `json:"fixtures"`
	DigestSeeds   []corpusDigestSeed `json:"digestSeeds"`
}

type corpusFixture struct {
	Package string `json:"package"`
	Path    string `json:"path"`
	SHA256  string `json:"sha256"`
}

type corpusDigestSeed struct {
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
}

func TestCompatibilityCorpusIsExactAndClosed(t *testing.T) {
	manifest := readCorpusManifest(t)
	if manifest.SchemaVersion != corpusManifestSchemaV1 || len(manifest.Fixtures) != 4 || len(manifest.DigestSeeds) != 1 {
		t.Fatalf("unexpected compatibility corpus identity: %+v", manifest)
	}
	packages := make([]string, 0, len(manifest.Fixtures))
	for _, fixture := range manifest.Fixtures {
		payload := readCompatibilityFixture(t, fixture)
		if err := decodeClosedFixture(fixture.Package, payload); err != nil {
			t.Fatalf("strictly decode %s: %v", fixture.Path, err)
		}
		var open map[string]any
		if err := json.Unmarshal(payload, &open); err != nil {
			t.Fatal(err)
		}
		open["unknown_contract_field"] = true
		unknown, err := json.Marshal(open)
		if err != nil {
			t.Fatal(err)
		}
		if err := decodeClosedFixture(fixture.Package, unknown); err == nil {
			t.Fatalf("%s accepted an unknown JSON field", fixture.Path)
		}
		packages = append(packages, fixture.Package)
	}
	sort.Strings(packages)
	want := []string{"providerexecutor/v1beta1", "runtimeinventory", "runtimelease", "stackaction"}
	if fmt.Sprint(packages) != fmt.Sprint(want) {
		t.Fatalf("corpus packages = %v, want %v", packages, want)
	}
	seed := manifest.DigestSeeds[0]
	if len(readCompatibilityFile(t, seed.Path, seed.SHA256)) == 0 {
		t.Fatal("digest corpus seed is empty")
	}
}

func FuzzClosedContractJSON(f *testing.F) {
	manifest := readCorpusManifest(f)
	for _, fixture := range manifest.Fixtures {
		f.Add(fixture.Package, readCompatibilityFixture(f, fixture))
	}
	f.Fuzz(func(t *testing.T, packageName string, payload []byte) {
		if !isContractPackage(packageName) {
			t.Skip()
		}
		if err := decodeClosedFixture(packageName, payload); err != nil {
			return
		}
		var value any
		if err := json.Unmarshal(payload, &value); err != nil {
			t.Fatalf("closed decoder accepted invalid JSON: %v", err)
		}
	})
}

func FuzzProviderDigestDeterminism(f *testing.F) {
	seed, err := os.ReadFile("compatibility/corpus/digest/native-ref.txt")
	if err != nil {
		f.Fatal(err)
	}
	f.Add(string(seed))
	f.Add("")
	f.Add("  provider-native://compat/server-1  ")
	canonical := regexp.MustCompile(`^sha256:[0-9a-f]{64}$`)
	f.Fuzz(func(t *testing.T, nativeRef string) {
		first := providerexecutor.ComputeNativeRefHash(nativeRef)
		second := providerexecutor.ComputeNativeRefHash(nativeRef)
		if first != second || !canonical.MatchString(first) {
			t.Fatalf("non-canonical deterministic digest %q / %q", first, second)
		}
	})
}

func readCorpusManifest(t testing.TB) corpusManifest {
	t.Helper()
	payload, err := os.ReadFile("compatibility/corpus/manifest.json")
	if err != nil {
		t.Fatal(err)
	}
	var manifest corpusManifest
	decoder := json.NewDecoder(bytes.NewReader(payload))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&manifest); err != nil {
		t.Fatal(err)
	}
	if err := requireJSONEOF(decoder); err != nil {
		t.Fatal(err)
	}
	return manifest
}

func readCompatibilityFixture(t testing.TB, fixture corpusFixture) []byte {
	t.Helper()
	return readCompatibilityFile(t, fixture.Path, fixture.SHA256)
}

func readCompatibilityFile(t testing.TB, path, expectedSHA256 string) []byte {
	t.Helper()
	clean := filepath.Clean(filepath.FromSlash(path))
	root := filepath.Clean(filepath.FromSlash("compatibility/corpus"))
	if clean == root || !strings.HasPrefix(clean, root+string(filepath.Separator)) {
		t.Fatalf("fixture escapes compatibility corpus: %q", path)
	}
	info, err := os.Lstat(clean)
	if err != nil || !info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0 || info.Size() > 1<<20 {
		t.Fatalf("fixture is not a bounded plain file: %q", path)
	}
	payload, err := os.ReadFile(clean)
	if err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(payload)
	if hex.EncodeToString(sum[:]) != expectedSHA256 {
		t.Fatalf("fixture digest drift for %s", path)
	}
	return payload
}

func decodeClosedFixture(packageName string, payload []byte) error {
	var target any
	switch packageName {
	case "providerexecutor/v1beta1":
		target = &providerexecutor.Command{}
	case "runtimeinventory":
		target = &runtimeinventory.ServerList{}
	case "runtimelease":
		target = &runtimelease.Lease{}
	case "stackaction":
		target = &stackaction.Request{}
	default:
		return fmt.Errorf("unknown contract package %q", packageName)
	}
	decoder := json.NewDecoder(bytes.NewReader(payload))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	return requireJSONEOF(decoder)
}

func requireJSONEOF(decoder *json.Decoder) error {
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		if err == nil {
			return fmt.Errorf("multiple JSON values")
		}
		return err
	}
	return nil
}

func isContractPackage(packageName string) bool {
	switch packageName {
	case "providerexecutor/v1beta1", "runtimeinventory", "runtimelease", "stackaction":
		return true
	default:
		return false
	}
}
