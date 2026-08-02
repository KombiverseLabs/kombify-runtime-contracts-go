package contracts_test

import (
	"os"
	"strings"
	"testing"
)

func TestDeliveryReleaseRequiresSuffixFreeNumericSemVer(t *testing.T) {
	t.Parallel()

	payload, err := os.ReadFile(".github/workflows/delivery-release.yml")
	if err != nil {
		t.Fatal(err)
	}
	workflow := string(payload)
	wantPattern := `[[ "$DELIVERY_VERSION" =~ ^(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)$ ]]`
	if !strings.Contains(workflow, wantPattern) {
		t.Fatal("delivery release does not enforce suffix-free numeric SemVer")
	}
	if strings.Contains(workflow, `-[0-9A-Za-z.-]+`) {
		t.Fatal("delivery release still accepts a prerelease suffix")
	}
	if !strings.Contains(workflow, `fast-pre-1.0) [[ "$DELIVERY_VERSION" == 0.* ]]`) {
		t.Fatal("fast pre-1.0 profile is not bound to numeric v0.x")
	}
}

func TestDeliveryPushIsCompileOnlyUntilConsumerCutover(t *testing.T) {
	t.Parallel()

	payload, err := os.ReadFile(".github/workflows/delivery.yml")
	if err != nil {
		t.Fatal(err)
	}
	workflow := string(payload)
	if !strings.Contains(workflow, `execute: ${{ github.event_name == 'workflow_dispatch' && inputs.execute }}`) {
		t.Fatal("delivery push could publish before coordinated consumer gates")
	}
}
