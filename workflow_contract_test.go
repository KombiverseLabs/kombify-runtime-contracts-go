package contracts_test

import (
	"encoding/json"
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
	if !strings.Contains(workflow, `go run ./internal/cmd/deliverytask -phase=publish`) {
		t.Fatal("delivery release does not validate the requested version and tag against .kombify/VERSION")
	}
}

func TestDeliveryReleaseRecompilesAndBindsExactPlan(t *testing.T) {
	t.Parallel()

	payload, err := os.ReadFile(".github/workflows/delivery-release.yml")
	if err != nil {
		t.Fatal(err)
	}
	workflow := string(payload)
	for _, required := range []string{
		`run-name: delivery-release-${{ inputs.delivery_release_id }}-${{ inputs.delivery_plan_digest }}`,
		`DELIVERY_STANDARDS_REF: ${{ inputs.delivery_standards_ref }}`,
		`TRUSTED_DELIVERY_STANDARDS_REF: ${{ vars.KOMBIFY_DELIVERY_STANDARDS_REF }}`,
		`[[ ! "$DELIVERY_STANDARDS_REF" =~ ^[0-9a-f]{40}$ ]]`,
		`[[ "$DELIVERY_STANDARDS_REF" != "$TRUSTED_DELIVERY_STANDARDS_REF" ]]`,
		`ref: ${{ inputs.delivery_standards_ref }}`,
		`test "$(git -C standards rev-parse HEAD)" = "$DELIVERY_STANDARDS_REF"`,
		`node standards/scripts/delivery-platform.mjs compile`,
		`node standards/scripts/delivery-platform.mjs validate-plan`,
		`plan_digest: process.env.DELIVERY_PLAN_DIGEST`,
		`release_id: process.env.DELIVERY_RELEASE_ID`,
		`plan.sources?.["kombify-runtime-contracts-go"] !== expected.source_sha`,
	} {
		if !strings.Contains(workflow, required) {
			t.Fatalf("delivery release does not bind the adapter to the exact compiled plan: missing %q", required)
		}
	}
}

func TestDeliveryReleaseIdempotencyRequiresExactMetadata(t *testing.T) {
	t.Parallel()

	payload, err := os.ReadFile(".github/workflows/delivery-release.yml")
	if err != nil {
		t.Fatal(err)
	}
	workflow := string(payload)
	for _, required := range []string{
		`--json tagName,isPrerelease,name,body,url`,
		`release.name !== process.env.EXPECTED_TITLE`,
		`release.body !== process.env.EXPECTED_BODY`,
		`release.isPrerelease !== expectedPrerelease`,
	} {
		if !strings.Contains(workflow, required) {
			t.Fatalf("delivery release can accept conflicting existing metadata: missing %q", required)
		}
	}
}

func TestDeliveryPlanUsesExactSourceVersion(t *testing.T) {
	t.Parallel()

	payload, err := os.ReadFile(".github/workflows/delivery.yml")
	if err != nil {
		t.Fatal(err)
	}
	workflow := string(payload)
	if !strings.Contains(workflow, `fs.readFileSync(".kombify/VERSION", "utf8").trim()`) ||
		!strings.Contains(workflow, `process.stdout.write(declared)`) {
		t.Fatal("delivery plan version is not derived from the exact source version")
	}
	if strings.Contains(workflow, `rev-list`) || strings.Contains(workflow, `BigInt(match[3])`) {
		t.Fatal("delivery plan still synthesizes an uncommitted release version")
	}
}

func TestDeliveryExecutionIsExplicitAndFailClosed(t *testing.T) {
	t.Parallel()

	payload, err := os.ReadFile(".github/workflows/delivery.yml")
	if err != nil {
		t.Fatal(err)
	}
	workflow := string(payload)
	if !strings.Contains(workflow, `EXECUTE_DELIVERY: ${{ github.event_name == 'workflow_dispatch' && inputs.execute }}`) {
		t.Fatal("delivery push or pull request could publish without an explicit dispatch")
	}
	if !strings.Contains(workflow, `REQUIRE_FAST_PREFLIGHT: "true"`) ||
		!strings.Contains(workflow, `FAIL_CLOSED_FAST: "true"`) ||
		!strings.Contains(workflow, `run_operation build`) ||
		!strings.Contains(workflow, `run_operation validate`) {
		t.Fatal("pre-1.0 delivery does not require build and validation before publication")
	}
	if !strings.Contains(workflow, `"$DELIVERY_PROFILE" == "fast-pre-1.0" && "$FAIL_CLOSED_FAST" != "true"`) {
		t.Fatal("pre-1.0 failure handling is not bound to the fail-closed public-repository mode")
	}
	if !strings.Contains(workflow, `STANDARDS_REF: ${{ vars.KOMBIFY_DELIVERY_STANDARDS_REF }}`) ||
		!strings.Contains(workflow, `[[ ! "$STANDARDS_REF" =~ ^[0-9a-f]{40}$ ]]`) ||
		!strings.Contains(workflow, `test "$(git -C standards rev-parse HEAD)" = "$STANDARDS_REF"`) ||
		!strings.Contains(workflow, `DELIVERY_STANDARDS_REF: process.env.STANDARDS_REF`) {
		t.Fatal("public delivery is not pinned to the exact classified compiler snapshot")
	}
	runtime, err := os.ReadFile(".kombify/delivery.mjs")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(runtime), `DELIVERY_STANDARDS_REF: requireString(`) ||
		!strings.Contains(string(runtime), `!/^[0-9a-f]{40}$/.test(variables.DELIVERY_STANDARDS_REF)`) {
		t.Fatal("repository runtime does not fail closed on a missing or invalid compiler ref")
	}

	operations, err := os.ReadFile(".kombify/delivery-operations.json")
	if err != nil {
		t.Fatal(err)
	}
	var bindings struct {
		Groups []struct {
			Operations struct {
				Publish struct {
					Fast []struct {
						WaitForCompletion *bool                      `json:"wait_for_completion"`
						RunNameContains   string                     `json:"run_name_contains"`
						Inputs            map[string]json.RawMessage `json:"inputs"`
					} `json:"fast"`
					Stable []struct {
						WaitForCompletion *bool                      `json:"wait_for_completion"`
						RunNameContains   string                     `json:"run_name_contains"`
						Inputs            map[string]json.RawMessage `json:"inputs"`
					} `json:"stable"`
				} `json:"publish"`
			} `json:"operations"`
		} `json:"groups"`
	}
	if err := json.Unmarshal(operations, &bindings); err != nil {
		t.Fatal(err)
	}
	if len(bindings.Groups) != 1 || len(bindings.Groups[0].Operations.Publish.Fast) != 1 || len(bindings.Groups[0].Operations.Publish.Stable) != 1 {
		t.Fatal("delivery publish binding is not singular")
	}
	for profile, publish := range map[string]struct {
		WaitForCompletion *bool
		RunNameContains   string
		Inputs            map[string]json.RawMessage
	}{
		"fast": {
			WaitForCompletion: bindings.Groups[0].Operations.Publish.Fast[0].WaitForCompletion,
			RunNameContains:   bindings.Groups[0].Operations.Publish.Fast[0].RunNameContains,
			Inputs:            bindings.Groups[0].Operations.Publish.Fast[0].Inputs,
		},
		"stable": {
			WaitForCompletion: bindings.Groups[0].Operations.Publish.Stable[0].WaitForCompletion,
			RunNameContains:   bindings.Groups[0].Operations.Publish.Stable[0].RunNameContains,
			Inputs:            bindings.Groups[0].Operations.Publish.Stable[0].Inputs,
		},
	} {
		if publish.WaitForCompletion == nil || !*publish.WaitForCompletion {
			t.Fatalf("%s delivery could report success before the exact release adapter completes", profile)
		}
		if _, exists := publish.Inputs["wait_for_completion"]; exists {
			t.Fatalf("%s delivery sends an input not declared by the release workflow", profile)
		}
		if publish.RunNameContains != "${DELIVERY_RELEASE_ID}-${DELIVERY_PLAN_DIGEST}" {
			t.Fatalf("%s delivery is not correlated by exact release and plan identity", profile)
		}
		if got := string(publish.Inputs["delivery_standards_ref"]); got != `"${DELIVERY_STANDARDS_REF}"` {
			t.Fatalf("%s delivery does not pass the exact compiler ref: %s", profile, got)
		}
	}
}
