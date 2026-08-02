// Package releaseversion validates the immutable relationship between source
// version metadata and a requested module release.
package releaseversion

import (
	"fmt"
	"os"
	"regexp"
	"strings"
)

var semverPattern = regexp.MustCompile(`^(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)$`)

// ReadSource reads a suffix-free SemVer from path. One trailing line ending is
// accepted; all other whitespace and additional lines are rejected.
func ReadSource(path string) (string, error) {
	payload, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read source version: %w", err)
	}
	version := string(payload)
	version = strings.TrimSuffix(version, "\n")
	version = strings.TrimSuffix(version, "\r")
	if !semverPattern.MatchString(version) {
		return "", fmt.Errorf("source version %q must be suffix-free MAJOR.MINOR.PATCH SemVer", version)
	}
	return version, nil
}

// ValidateRequested binds a requested release version and tag to the exact
// version committed in source.
func ValidateRequested(sourceVersion, requestedVersion, requestedTag string) error {
	if !semverPattern.MatchString(sourceVersion) {
		return fmt.Errorf("source version %q must be suffix-free MAJOR.MINOR.PATCH SemVer", sourceVersion)
	}
	if !semverPattern.MatchString(requestedVersion) {
		return fmt.Errorf("requested version %q must be suffix-free MAJOR.MINOR.PATCH SemVer", requestedVersion)
	}
	if requestedVersion != sourceVersion {
		return fmt.Errorf("requested version %q does not match source version %q", requestedVersion, sourceVersion)
	}
	wantTag := "v" + sourceVersion
	if requestedTag != wantTag {
		return fmt.Errorf("requested tag %q does not match source tag %q", requestedTag, wantTag)
	}
	return nil
}
