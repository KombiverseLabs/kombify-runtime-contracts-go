package releaseversion

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReadSource(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name    string
		payload string
		want    string
		wantErr string
	}{
		{name: "bare", payload: "0.1.4", want: "0.1.4"},
		{name: "line feed", payload: "0.1.4\n", want: "0.1.4"},
		{name: "windows line ending", payload: "0.1.4\r\n", want: "0.1.4"},
		{name: "prerelease", payload: "0.1.4-rc.1\n", wantErr: "suffix-free"},
		{name: "leading whitespace", payload: " 0.1.4\n", wantErr: "suffix-free"},
		{name: "extra line", payload: "0.1.4\n0.1.5\n", wantErr: "suffix-free"},
	} {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			path := filepath.Join(t.TempDir(), "VERSION")
			if err := os.WriteFile(path, []byte(test.payload), 0o600); err != nil {
				t.Fatal(err)
			}
			got, err := ReadSource(path)
			if test.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), test.wantErr) {
					t.Fatalf("ReadSource() error = %v, want containing %q", err, test.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if got != test.want {
				t.Fatalf("ReadSource() = %q, want %q", got, test.want)
			}
		})
	}
}

func TestValidateRequested(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name    string
		source  string
		version string
		tag     string
		wantErr string
	}{
		{name: "exact", source: "0.1.4", version: "0.1.4", tag: "v0.1.4"},
		{name: "version mismatch", source: "0.1.4", version: "0.1.3", tag: "v0.1.3", wantErr: "does not match source version"},
		{name: "tag mismatch", source: "0.1.4", version: "0.1.4", tag: "v0.1.3", wantErr: "does not match source tag"},
		{name: "missing tag", source: "0.1.4", version: "0.1.4", wantErr: "does not match source tag"},
		{name: "invalid source", source: "v0.1.4", version: "0.1.4", tag: "v0.1.4", wantErr: "source version"},
		{name: "prerelease request", source: "0.1.4", version: "0.1.4-rc.1", tag: "v0.1.4-rc.1", wantErr: "requested version"},
	} {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			err := ValidateRequested(test.source, test.version, test.tag)
			if test.wantErr == "" {
				if err != nil {
					t.Fatal(err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), test.wantErr) {
				t.Fatalf("ValidateRequested() error = %v, want containing %q", err, test.wantErr)
			}
		})
	}
}
