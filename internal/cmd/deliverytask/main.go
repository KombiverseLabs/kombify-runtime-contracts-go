// Command deliverytask validates side-effect-free library delivery phases.
// Immutable tag and GitHub Release creation belongs to delivery-release.yml.
package main

import (
	"flag"
	"fmt"
	"os"
	"regexp"
	"strings"
)

var semverPattern = regexp.MustCompile(`^(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)(-[0-9A-Za-z.-]+)?$`)

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
	version := strings.TrimSpace(os.Getenv("DELIVERY_VERSION"))
	if version == "" {
		payload, err := os.ReadFile(".kombify/VERSION")
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		version = strings.TrimSpace(string(payload))
	}
	if !semverPattern.MatchString(version) {
		fmt.Fprintf(os.Stderr, "invalid delivery version %q\n", version)
		os.Exit(1)
	}
	if *phase == "promote" && strings.HasPrefix(version, "0.") {
		fmt.Fprintln(os.Stderr, "pre-1.0 modules cannot use stable promotion")
		os.Exit(1)
	}
	fmt.Printf("validated %s metadata for v%s; immutable publication remains in delivery-release.yml\n", *phase, version)
}
