package dodeval_test

import (
	"testing"

	"github.com/buchin/dodeval"
)

const sampleMarkdown = `
# Task Specification

Berikut adalah rincian tugas untuk sprint ini:

### Definition of Done
- [x] Working tree clean dan tidak ada uncommitted files
- [ ] Targeted tests pass 100% dalam waktu kurang dari 10 detik
- [ ] Kompilasi build Go 0 error
- [ ] Zero context overload pada runner dengan micro-package terisolasi
`

func TestParseChecklist_MarkdownCheckboxes(t *testing.T) {
	items := dodeval.ParseChecklist(sampleMarkdown)
	if len(items) != 4 {
		t.Fatalf("expected 4 criteria, got %d", len(items))
	}

	if !items[0].Completed {
		t.Errorf("expected first item to be completed (- [x])")
	}
	if items[1].Completed {
		t.Errorf("expected second item to be uncompleted (- [ ])")
	}
	if items[0].Text != "Working tree clean dan tidak ada uncommitted files" {
		t.Errorf("unexpected text for first item: %s", items[0].Text)
	}
}

func TestParseChecklist_NumberedList(t *testing.T) {
	numbered := `
## Acceptance Criteria
1. Single Responsibility Principle: each module has its own domain
2. Zero Context Overload: runner focuses on small files
3. Reusability: 100% unit testable with zero dependencies
`
	items := dodeval.ParseChecklist(numbered)
	if len(items) != 3 {
		t.Fatalf("expected 3 items, got %d", len(items))
	}
	if items[0].Completed {
		t.Errorf("numbered items without checkbox should default to completed: false")
	}
	if items[0].Text != "Single Responsibility Principle: each module has its own domain" {
		t.Errorf("unexpected text: %s", items[0].Text)
	}
}

func TestEvaluator_AllCompletedPrechecked(t *testing.T) {
	checklist := `
- [x] Criterion 1
- [x] Criterion 2
`
	evaluator := dodeval.NewEvaluator()
	res := evaluator.EvaluateString(checklist, dodeval.Evidence{})

	if !res.Passed {
		t.Fatalf("expected evaluation to pass")
	}
	if res.CompletedCriteria != 2 || res.TotalCriteria != 2 {
		t.Errorf("expected 2/2 completed, got %d/%d", res.CompletedCriteria, res.TotalCriteria)
	}
	if res.Percentage != 100.0 {
		t.Errorf("expected 100%%, got %.1f%%", res.Percentage)
	}
	if res.Action != "mark_done" {
		t.Errorf("expected action mark_done, got %s", res.Action)
	}
}

func TestEvaluator_EvidenceFulfillment(t *testing.T) {
	checklist := `
- [ ] Targeted tests pass 100% dalam waktu kurang dari 10 detik
- [ ] Kompilasi build Go 0 error
- [ ] Dokumentasi dan Landing Page diperbarui
`
	evaluator := dodeval.NewEvaluator()

	// 1. Partial evidence: only tests and build reported
	partialEvidence := dodeval.Evidence{
		Summary:     "Kompilasi build Go 0 error berhasil diverifikasi.",
		TestSummary: "Targeted tests pass 100% dalam waktu 0.05 detik.",
	}

	res1 := evaluator.EvaluateString(checklist, partialEvidence)
	if res1.Passed {
		t.Fatalf("partial evidence should not pass 100%% DoD")
	}
	if res1.CompletedCriteria != 2 {
		t.Errorf("expected 2 completed criteria, got %d", res1.CompletedCriteria)
	}
	if len(res1.Gaps) != 1 {
		t.Errorf("expected 1 gap, got %d", len(res1.Gaps))
	}
	if res1.Action != "auto_requeue" {
		t.Errorf("expected action auto_requeue, got %s", res1.Action)
	}

	// 2. Full evidence: includes documentation
	fullEvidence := dodeval.Evidence{
		Summary:     "Kompilasi build Go 0 error dan dokumentasi serta Landing Page diperbarui.",
		TestSummary: "Targeted tests pass 100% dalam waktu 0.05 detik.",
		Assertions: []string{
			"Dokumentasi dan Landing Page diperbarui dengan kapabilitas terbaru",
		},
	}

	res2 := evaluator.EvaluateString(checklist, fullEvidence)
	if !res2.Passed {
		t.Fatalf("full evidence should pass: gaps=%v", res2.GapDescriptions)
	}
	if res2.CompletedCriteria != 3 {
		t.Errorf("expected 3 completed criteria, got %d", res2.CompletedCriteria)
	}
	if res2.Action != "mark_done" {
		t.Errorf("expected action mark_done, got %s", res2.Action)
	}
}

func TestEvaluator_ExplicitMetadataVerification(t *testing.T) {
	checklist := `
- [ ] Special hardware security key verification
`
	evaluator := dodeval.NewEvaluator()

	ev := dodeval.Evidence{
		Metadata: map[string]any{
			"verified_criteria": []string{"Special hardware security key verification"},
		},
	}

	res := evaluator.EvaluateString(checklist, ev)
	if !res.Passed {
		t.Fatalf("expected explicit metadata verification to satisfy criterion")
	}
}

func TestEvaluator_EmptyChecklist(t *testing.T) {
	evaluator := dodeval.NewEvaluator()
	res := evaluator.EvaluateString("", dodeval.Evidence{})

	if !res.Passed {
		t.Fatalf("empty checklist should pass by default")
	}
	if res.Action != "mark_done" {
		t.Errorf("expected action mark_done, got %s", res.Action)
	}
}

func TestParseAcceptanceCriteriaAndAnalyzeGaps(t *testing.T) {
	markdown := `
## Acceptance Criteria
1. [ ] Verifikasi unit test mandiri pada micro-packages pkg/taskfsm dan pkg/dodeval (100% pass)
2. [ ] Selesaikan Phase 3: Integrasi micro-packages ke task runner Cuan secara backward-compatible
`
	criteria := dodeval.ParseAcceptanceCriteria(markdown)
	if len(criteria) != 2 {
		t.Fatalf("expected 2 criteria, got %d", len(criteria))
	}

	ev := dodeval.Evidence{
		Summary:           "Selesaikan Phase 3: Integrasi micro-packages ke task runner Cuan secara backward-compatible",
		CompleteNote:      "Verifikasi unit test mandiri pada micro-packages pkg/taskfsm dan pkg/dodeval (100% pass)",
		CleanWorkingTree:  true,
		TargetedTestsPass: true,
		BuildZeroError:    true,
	}

	res := dodeval.AnalyzeGaps(criteria, ev)
	if !res.Passed {
		t.Fatalf("expected AnalyzeGaps to pass with full evidence, got %d gaps: %v", len(res.Gaps), res.GapDescriptions)
	}
}

func BenchmarkParseChecklist(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = dodeval.ParseChecklist(sampleMarkdown)
	}
}

func BenchmarkEvaluate(b *testing.B) {
	evaluator := dodeval.NewEvaluator()
	ev := dodeval.Evidence{
		Summary:     "Targeted tests pass 100% and build 0 error",
		TestSummary: "PASS 0.01s",
	}
	criteria := dodeval.ParseChecklist(sampleMarkdown)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = evaluator.Evaluate(criteria, ev)
	}
}
