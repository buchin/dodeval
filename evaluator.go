package dodeval

import (
	"fmt"
	"strings"
	"time"
)

// Evidence holds reported facts, test outputs, notes and artifact references.
// The package does not independently verify the submitted information.
type Evidence struct {
	Assertions        []string       `json:"assertions,omitempty"`
	Summary           string         `json:"summary,omitempty"`
	Notes             string         `json:"notes,omitempty"`
	CompleteNote      string         `json:"complete_note,omitempty"`
	Artifacts         []string       `json:"artifacts,omitempty"`
	TestSummary       string         `json:"test_summary,omitempty"`
	Metadata          map[string]any `json:"metadata,omitempty"`
	CleanWorkingTree  bool           `json:"clean_working_tree"`
	TargetedTestsPass bool           `json:"targeted_tests_pass"`
	BuildZeroError    bool           `json:"build_zero_error"`
	TestOutput        string         `json:"test_output,omitempty"`
	CommitSHA         string         `json:"commit_sha,omitempty"`
}

// EvaluationResult contains a heuristic comparison of criteria and reported text.
type EvaluationResult struct {
	Passed            bool          `json:"passed"`
	TotalCriteria     int           `json:"total_criteria"`
	CompletedCriteria int           `json:"completed_criteria"`
	Percentage        float64       `json:"percentage"`
	Completed         []Criterion   `json:"completed"`
	Gaps              []Criterion   `json:"gaps,omitempty"`
	GapDescriptions   []string      `json:"gap_descriptions,omitempty"`
	Summary           string        `json:"summary"`
	Action            string        `json:"action"` // "mark_done" | "auto_requeue" | "review"
	Duration          time.Duration `json:"duration"`
}

// EvaluatorOption configures evaluator behavior.
type EvaluatorOption func(*Evaluator)

// WithKeywordThreshold sets the fraction of keywords required to match evidence (0.0 to 1.0).
func WithKeywordThreshold(threshold float64) EvaluatorOption {
	return func(e *Evaluator) {
		if threshold > 0 && threshold <= 1.0 {
			e.keywordThreshold = threshold
		}
	}
}

// WithMinPassPercentage sets required passing percentage (default 100.0).
func WithMinPassPercentage(pct float64) EvaluatorOption {
	return func(e *Evaluator) {
		if pct > 0 && pct <= 100.0 {
			e.minPassPercentage = pct
		}
	}
}

// Evaluator performs gap analysis comparing criteria against reported evidence.
type Evaluator struct {
	keywordThreshold  float64
	minPassPercentage float64
}

// NewEvaluator creates a new Evaluator with defaults.
func NewEvaluator(opts ...EvaluatorOption) *Evaluator {
	e := &Evaluator{
		keywordThreshold:  0.60,
		minPassPercentage: 100.0,
	}
	for _, opt := range opts {
		opt(e)
	}
	return e
}

// Evaluate evaluates a slice of Criterion against provided Evidence.
func (e *Evaluator) Evaluate(criteria []Criterion, evidence Evidence) *EvaluationResult {
	start := time.Now()

	res := &EvaluationResult{
		TotalCriteria: len(criteria),
		Completed:     make([]Criterion, 0),
		Gaps:          make([]Criterion, 0),
	}

	if len(criteria) == 0 {
		res.Passed = true
		res.Percentage = 100.0
		res.Action = "mark_done"
		res.Summary = "Tidak ada kriteria spesifik yang disyaratkan; evaluasi otomatis lulus."
		res.Duration = time.Since(start)
		return res
	}

	// Prepare searchable corpus from evidence
	corpus := buildEvidenceCorpus(evidence)

	for _, c := range criteria {
		crit := c
		if crit.Completed {
			res.Completed = append(res.Completed, crit)
			continue
		}

		if e.matchesEvidence(crit, corpus, evidence) {
			crit.Completed = true
			res.Completed = append(res.Completed, crit)
		} else {
			res.Gaps = append(res.Gaps, crit)
			res.GapDescriptions = append(res.GapDescriptions, fmt.Sprintf("Kriteria '%s' belum terverifikasi oleh bukti laporan", crit.Text))
		}
	}

	res.CompletedCriteria = len(res.Completed)
	res.Percentage = (float64(res.CompletedCriteria) / float64(res.TotalCriteria)) * 100.0
	res.Passed = res.Percentage >= e.minPassPercentage

	if res.Passed {
		res.Action = "mark_done"
		res.Summary = fmt.Sprintf("Evaluasi DoD LULUS: %d/%d (%.1f%%) kriteria terpenuhi dengan bukti lengkap.",
			res.CompletedCriteria, res.TotalCriteria, res.Percentage)
	} else {
		res.Action = "auto_requeue"
		res.Summary = fmt.Sprintf("Evaluasi DoD BELUM LULUS: %d/%d (%.1f%%) kriteria terpenuhi. Masih ada %d gap yang perlu dipenuhi.",
			res.CompletedCriteria, res.TotalCriteria, res.Percentage, len(res.Gaps))
	}

	res.Duration = time.Since(start)
	return res
}

// EvaluateString parses markdown checklist from string and evaluates against evidence.
func (e *Evaluator) EvaluateString(checklistMarkdown string, evidence Evidence) *EvaluationResult {
	criteria := ParseChecklist(checklistMarkdown)
	return e.Evaluate(criteria, evidence)
}

// buildEvidenceCorpus combines all textual evidence fields into a lowercase normalized string.
func buildEvidenceCorpus(ev Evidence) string {
	var sb strings.Builder
	for _, a := range ev.Assertions {
		sb.WriteString(strings.ToLower(a))
		sb.WriteString(" ")
	}
	sb.WriteString(strings.ToLower(ev.Summary))
	sb.WriteString(" ")
	sb.WriteString(strings.ToLower(ev.Notes))
	sb.WriteString(" ")
	sb.WriteString(strings.ToLower(ev.CompleteNote))
	sb.WriteString(" ")
	sb.WriteString(strings.ToLower(ev.TestSummary))
	sb.WriteString(" ")
	sb.WriteString(strings.ToLower(ev.TestOutput))
	sb.WriteString(" ")
	sb.WriteString(strings.ToLower(ev.CommitSHA))
	sb.WriteString(" ")
	for _, art := range ev.Artifacts {
		sb.WriteString(strings.ToLower(art))
		sb.WriteString(" ")
	}
	if ev.CleanWorkingTree {
		sb.WriteString(" working tree clean git clean status ")
	}
	if ev.TargetedTestsPass {
		sb.WriteString(" targeted automated tests pass 100% unit tests pass test pass ")
	}
	if ev.BuildZeroError {
		sb.WriteString(" build zero error kompilasi 0 error build pass ")
	}
	return sb.String()
}

// AnalyzeGaps evaluates criteria against evidence using the default Evaluator.
func AnalyzeGaps(criteria []Criterion, evidence Evidence) *EvaluationResult {
	return NewEvaluator().Evaluate(criteria, evidence)
}

// matchesEvidence checks if a criterion is satisfied by the evidence corpus.
func (e *Evaluator) matchesEvidence(c Criterion, corpus string, ev Evidence) bool {
	// 1. Direct substring match of clean text
	lowerText := strings.ToLower(c.Text)
	if strings.Contains(corpus, lowerText) {
		return true
	}

	// 2. Metadata explicit verification list (e.g. verified_ids or verified_criteria)
	if ev.Metadata != nil {
		if rawList, ok := ev.Metadata["verified_criteria"]; ok {
			if list, ok := rawList.([]string); ok {
				for _, v := range list {
					if strings.EqualFold(v, c.ID) || strings.EqualFold(v, c.Text) {
						return true
					}
				}
			}
		}
	}

	// 3. Keyword density match
	if len(c.Keywords) == 0 {
		return false
	}

	matchedKeywords := 0
	for _, kw := range c.Keywords {
		if strings.Contains(corpus, kw) {
			matchedKeywords++
		}
	}

	ratio := float64(matchedKeywords) / float64(len(c.Keywords))
	return ratio >= e.keywordThreshold
}
