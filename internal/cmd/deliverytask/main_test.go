package main

import "testing"

func TestDeliveryVersionRequiresSuffixFreeNumericSemVer(t *testing.T) {
	t.Parallel()

	for _, valid := range []string{"0.1.0", "1.0.0", "12.34.56"} {
		if !semverPattern.MatchString(valid) {
			t.Errorf("expected %q to be accepted", valid)
		}
	}
	for _, invalid := range []string{"v0.1.0", "0.1.0-dev.1", "0.1.0+build", "01.1.0", "0.01.0", "0.1"} {
		if semverPattern.MatchString(invalid) {
			t.Errorf("expected %q to be rejected", invalid)
		}
	}
}
