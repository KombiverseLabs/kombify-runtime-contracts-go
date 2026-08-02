// Command deliverytask validates side-effect-free library delivery phases.
// Immutable tag and GitHub Release creation belongs to delivery-release.yml.
package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/KombiverseLabs/kombify-runtime-contracts-go/internal/releaseversion"
)

func main() {
	phase := flag.String("phase", "", "delivery phase")
	flag.Parse()
	if *phase != "publish" && *phase != "promote" && *phase != "async" {
		fmt.Fprintf(os.Stderr, "unsupported delivery phase %q\n", *phase)
		os.Exit(2)
	}
	if *phase == "async" {
		fmt.Println("no asynchronous delivery phase for a Go contract module")
		return
	}
	sourceVersion, err := releaseversion.ReadSource(".kombify/VERSION")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	version := os.Getenv("DELIVERY_VERSION")
	if version == "" {
		version = sourceVersion
	}
	tag := os.Getenv("DELIVERY_TAG")
	if tag == "" {
		tag = "v" + sourceVersion
	}
	if err := releaseversion.ValidateRequested(sourceVersion, version, tag); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if *phase == "promote" && strings.HasPrefix(version, "0.") {
		fmt.Fprintln(os.Stderr, "pre-1.0 modules cannot use stable promotion")
		os.Exit(1)
	}
	fmt.Printf("validated %s metadata for v%s; immutable publication remains in delivery-release.yml\n", *phase, version)
}
