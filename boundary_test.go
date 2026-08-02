package contracts_test

import (
	"bytes"
	"encoding/json"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

var publicPackages = []string{
	"runtimelease",
	"providerexecutor/v1beta1",
	"runtimeinventory",
	"stackaction",
}

func TestRepositoryBoundary(t *testing.T) {
	t.Parallel()

	allowedPublicDirs := make(map[string]bool, len(publicPackages))
	for _, packagePath := range publicPackages {
		allowedPublicDirs[filepath.Clean(filepath.FromSlash(packagePath))] = true
	}
	err := filepath.WalkDir(".", func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			if path == ".git" || path == "vendor" || path == "internal" {
				return filepath.SkipDir
			}
			return nil
		}
		if filepath.Ext(path) != ".go" || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		dir := filepath.Clean(filepath.Dir(path))
		if !allowedPublicDirs[dir] {
			return &boundaryError{path: path, message: "unexpected public Go package directory " + dir}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	for _, packagePath := range publicPackages {
		info, err := os.Stat(filepath.FromSlash(packagePath))
		if err != nil || !info.IsDir() {
			t.Fatalf("required public package %q is missing", packagePath)
		}
	}

	for _, retired := range []string{"vmlease", "runtimeaction", "serverruntime"} {
		if _, err := os.Stat(retired); !os.IsNotExist(err) {
			t.Fatalf("retired compatibility package %q must not exist", retired)
		}
	}

	module, err := os.ReadFile("go.mod")
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(module, []byte("github.com/KombiverseLabs/kombify-go-common")) {
		t.Fatal("public module must not depend on kombify-go-common")
	}
	rootVersion, err := os.ReadFile("VERSION")
	if err != nil {
		t.Fatal(err)
	}
	deliveryVersion, err := os.ReadFile(filepath.FromSlash(".kombify/VERSION"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(string(rootVersion)) != strings.TrimSpace(string(deliveryVersion)) {
		t.Fatal("VERSION and .kombify/VERSION must match")
	}

	forbiddenImports := []string{
		"github.com/KombiverseLabs/",
		"github.com/aws/",
		"github.com/aws/aws-sdk-go",
		"github.com/hetznercloud/",
		"github.com/ionos-cloud/",
		"github.com/linode/",
		"github.com/digitalocean/",
		"github.com/vultr/",
	}
	forbiddenFields := map[string]struct{}{
		"AccessKeyID": {}, "AgentToken": {}, "ClientPrivateKey": {},
		"KeyPEM": {}, "Password": {}, "PrivateKey": {},
		"RepoPassword": {}, "SecretAccessKey": {}, "Token": {},
	}

	for _, packagePath := range publicPackages {
		err := filepath.WalkDir(filepath.FromSlash(packagePath), func(path string, entry os.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if entry.IsDir() || filepath.Ext(path) != ".go" || strings.HasSuffix(path, "_test.go") {
				return nil
			}
			file, parseErr := parser.ParseFile(token.NewFileSet(), path, nil, parser.ImportsOnly|parser.SkipObjectResolution)
			if parseErr != nil {
				return parseErr
			}
			for _, imported := range file.Imports {
				value, unquoteErr := strconv.Unquote(imported.Path.Value)
				if unquoteErr != nil {
					return unquoteErr
				}
				for _, prefix := range forbiddenImports {
					if strings.HasPrefix(value, prefix) {
						return &boundaryError{path: path, message: "forbidden public import " + value}
					}
				}
			}

			full, fullErr := parser.ParseFile(token.NewFileSet(), path, nil, parser.SkipObjectResolution)
			if fullErr != nil {
				return fullErr
			}
			ast.Inspect(full, func(node ast.Node) bool {
				field, ok := node.(*ast.Field)
				if !ok {
					return true
				}
				for _, name := range field.Names {
					if _, forbidden := forbiddenFields[name.Name]; forbidden {
						walkErr = &boundaryError{path: path, message: "secret-bearing field " + name.Name}
						return false
					}
				}
				return true
			})
			return walkErr
		})
		if err != nil {
			t.Fatal(err)
		}
	}

	err = filepath.WalkDir(".", func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			if path == ".git" || path == "vendor" {
				return filepath.SkipDir
			}
			return nil
		}
		if filepath.Ext(path) != ".json" {
			return nil
		}
		payload, readErr := os.ReadFile(path)
		if readErr != nil {
			return readErr
		}
		var value any
		if jsonErr := json.Unmarshal(payload, &value); jsonErr != nil {
			return &boundaryError{path: path, message: "invalid Golden JSON: " + jsonErr.Error()}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

type boundaryError struct {
	path    string
	message string
}

func (err *boundaryError) Error() string {
	return err.path + ": " + err.message
}
