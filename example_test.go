package dodeval_test

import (
	"fmt"
	"github.com/buchin/dodeval"
)

func ExampleParseRequirements() {
	criteria := dodeval.ParseRequirements("DoD:\n- [x] Verify production health\n- [ ] Verify the requested access")
	for _, criterion := range criteria {
		fmt.Println(criterion.Text, criterion.Completed)
	}
	// Output:
	// Verify production health false
	// Verify the requested access false
}
