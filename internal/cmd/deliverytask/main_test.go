package main

import (
	"testing"

	"github.com/KombiverseLabs/kombify-runtime-contracts-go/internal/releaseversion"
)

func TestDeliveryVersionRequiresSuffixFreeNumericSemVer(t *testing.T) {
	t.Parallel()

	for _, valid := range []string{"0.1.0", "1.0.0", "12.34.56"} {
		if err := releaseversion.ValidateRequested(valid, valid, "v"+valid); err != nil {
			t.Errorf("expected %q to be accepted: %v", valid, err)
		}
	}
	for _, invalid := range []string{"v0.1.0", "0.1.0-dev.1", "0.1.0+build", "01.1.0", "0.01.0", "0.1"} {
		if err := releaseversion.ValidateRequested(invalid, invalid, "v"+invalid); err == nil {
			t.Errorf("expected %q to be rejected", invalid)
		}
	}
}
