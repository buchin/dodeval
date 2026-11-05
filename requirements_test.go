package dodeval_test

import (
	"github.com/buchin/dodeval"
	"strings"
	"testing"
)

func TestParseRequirementsNeverInheritsPlanCompletion(t *testing.T) {
	input := "- [x] Verify the exact account\n1) [X] Verify the exact product\n+ [ ] Verify access\n- verify access"
	items := dodeval.ParseRequirements(input)
	if len(items) != 3 {
		t.Fatalf("wrong requirement count: %d", len(items))
	}
	for _, item := range items {
		if item.Completed || !item.Required || item.Text == "" {
			t.Fatal("plan became verified evidence", item)
		}
	}
	tracked := dodeval.ParseChecklist("- [x] Verify access")
	if len(tracked) != 1 || !tracked[0].Completed {
		t.Fatal("tracking API changed")
	}
}
func TestParseRequirementsPrefersExplicitAcceptanceSection(t *testing.T) {
	input := "## Implementation\n- [x] Draft internal notes\n## Acceptance Criteria\n1. Verify the result\n- [X] Verify the result\n## Notes\n- Reporting procedure"
	items := dodeval.ParseRequirements(input)
	if len(items) != 1 || items[0].Text != "Verify the result" || items[0].Completed {
		t.Fatal("noncriteria became acceptance requirements", items)
	}
}
func TestParseRequirementsWithoutAcceptanceHeadingKeepsRequestedLists(t *testing.T) {
	for _, input := range []string{"## Approved scope\n1) Verify result", "DoD:\n- Verify result", "**Kriteria penerimaan:**\n- [x] Verify result"} {
		items := dodeval.ParseRequirements(input)
		if len(items) != 1 || items[0].Text != "Verify result" || items[0].Completed {
			t.Fatal(input, items)
		}
	}
}
func TestParseRequirementsDoesNotSilentlyLoseLongPreamble(t *testing.T) {
	items := dodeval.ParseRequirements(strings.Repeat("context ", 10000) + "\n- Verify result")
	if len(items) != 1 {
		t.Fatal("long preamble hid later requirement")
	}
}
