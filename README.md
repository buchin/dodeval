# dodeval

A dependency-free Go parser for acceptance requirements and textual gap analysis.

## Install

```sh
go get github.com/buchin/dodeval@v0.1.0
```

Use the Go version declared in [go.mod](go.mod) or a newer compatible version.

## Parse requested work

```go
package main

import (
    "fmt"
    "github.com/buchin/dodeval"
)

func main() {
    criteria := dodeval.ParseRequirements("DoD:\n- [x] Verify production health")
    fmt.Println(criteria[0].Text)      // Verify production health
    fmt.Println(criteria[0].Completed) // false: a plan is not verification
}
```

`ParseRequirements` accepts Markdown checkboxes, bullet lists, and numbered lists
using `1.` or `1)`. It prefers explicit acceptance sections when present, otherwise
reads the requested lists. English and Indonesian acceptance headings are
supported. Duplicate requirements are collapsed and every requirement starts
unverified, regardless of its original checkbox state. Long preambles do not
silently hide later requirements.

`ParseChecklist` and its `ParseAcceptanceCriteria` alias retain checkbox states
for checklist-tracking applications. Their existing behavior is preserved. Use
`ParseRequirements` when ingesting an approved plan or agent instructions.

## Analyze reported text

`AnalyzeGaps(criteria, evidence)` and `NewEvaluator` compare requirements with the
provided `Evidence` text and optional explicit metadata. Matching uses substrings
and configurable keyword overlap. It does not execute tests, inspect production,
validate artifacts, detect every contradiction, or establish that external
operations occurred. A checked item supplied to the tracking evaluator is trusted
as caller-provided state.

`EvaluationResult.Passed` describes that text comparison. The legacy `Action`
values (`mark_done`, `auto_requeue`, `review`) are recommendations, never executed
commands or authorization. Application approval and independent verification must
remain with the caller. An empty requirement set means no criteria were compared;
it does not establish that the original work was completed.

The package has no database, provider, process runner or application policy.

## Check

```sh
go test -race ./...
go vet ./...
```

The [executable example](example_test.go) is included in the tests.
[MIT license](LICENSE).
