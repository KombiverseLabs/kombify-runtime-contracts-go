// Command contractcheck performs repository-local contract checks without
// requiring a platform-private helper binary.
package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	mode := flag.String("mode", "format", "check mode")
	flag.Parse()
	if *mode != "format" {
		fmt.Fprintf(os.Stderr, "unsupported contractcheck mode %q\n", *mode)
		os.Exit(2)
	}

	var unformatted []string
	err := filepath.WalkDir(".", func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			if path == ".git" || path == "vendor" {
				return filepath.SkipDir
			}
			return nil
		}
		if filepath.Ext(path) != ".go" {
			return nil
		}
		source, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		formatted, err := format.Source(source)
		if err != nil {
			return fmt.Errorf("format %s: %w", path, err)
		}
		if !bytes.Equal(source, formatted) {
			unformatted = append(unformatted, filepath.ToSlash(path))
		}
		return nil
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if len(unformatted) != 0 {
		fmt.Fprintf(os.Stderr, "unformatted Go files:\n%s\n", strings.Join(unformatted, "\n"))
		os.Exit(1)
	}
}
