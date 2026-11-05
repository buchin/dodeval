package dodeval

import (
	"bufio"
	"fmt"
	"regexp"
	"strings"
)

// Criterion represents a single Definition of Done or Acceptance Criterion requirement.
type Criterion struct {
	ID        string   `json:"id"`
	Index     int      `json:"index"`
	Raw       string   `json:"raw"`
	Text      string   `json:"text"`
	Completed bool     `json:"completed"`
	Required  bool     `json:"required"`
	Keywords  []string `json:"keywords,omitempty"`
}

var (
	// regex for markdown task checklist items: - [ ] or - [x] or * [ ] or 1. [ ]
	taskCheckboxRegex = regexp.MustCompile(`^(?:\s*[-*+]|\s*\d+[.)])\s*\[([ xX])\]\s*(.*)$`)
	// regex for numbered or bullet list items without explicit checkbox: 1. text or - text
	listItemRegex = regexp.MustCompile(`^(?:\s*[-*+]|\s*\d+[.)])\s*(.*)$`)
)

// ParseChecklist parses markdown content or structured criteria text into a slice of Criterion.
func ParseChecklist(input string) []Criterion { return parseChecklist(input, false) }

// ParseRequirements extracts requested outcomes, never completed work. Explicit
// acceptance sections take precedence; otherwise ordinary lists are accepted.
// Checked plan boxes remain unverified and duplicate requirements are collapsed.
func ParseRequirements(input string) []Criterion { return parseChecklist(input, true) }

var acceptanceHeading = regexp.MustCompile(`(?i)^(?:definition of done|acceptance criteria|kriteria penerimaan|kriteria selesai|kriteria|dod)\s*[:—-]?\s*$`)

func isAcceptanceHeading(line string) bool {
	return acceptanceHeading.MatchString(strings.Trim(cleanMarkdownFormatting(line), "# *_\t"))
}

func parseChecklist(input string, requirements bool) []Criterion {
	var items []Criterion
	scanner := bufio.NewScanner(strings.NewReader(input))
	index := 0
	seen := map[string]bool{}
	if requirements {
		scanner.Buffer(make([]byte, 4096), max(64<<10, len(input)+1))
	}

	inCriteriaSection := false
	hasSectionHeaders := strings.Contains(input, "### ") || strings.Contains(input, "## ")
	if requirements {
		hasSectionHeaders = false
		for _, line := range strings.Split(input, "\n") {
			if isAcceptanceHeading(strings.TrimSpace(line)) {
				hasSectionHeaders = true
				break
			}
		}
	}

	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		lower := strings.ToLower(trimmed)
		if requirements {
			if isAcceptanceHeading(trimmed) {
				inCriteriaSection = true
				continue
			}
			if strings.HasPrefix(trimmed, "#") || (hasSectionHeaders && strings.HasSuffix(trimmed, ":") && !listItemRegex.MatchString(trimmed)) {
				inCriteriaSection = false
				continue
			}
		}
		if !requirements && strings.HasPrefix(lower, "#") {
			if strings.Contains(lower, "definition of done") ||
				strings.Contains(lower, "acceptance criteria") ||
				strings.Contains(lower, "kriteria selesai") ||
				strings.Contains(lower, "dod") {
				inCriteriaSection = true
				continue
			} else if inCriteriaSection {
				// Entered another major section header
				inCriteriaSection = false
			}
		}

		// If document has section headers, prefer parsing items inside criteria section
		// unless none found or line is an explicit markdown checkbox
		isCheckbox := taskCheckboxRegex.MatchString(trimmed)
		if hasSectionHeaders && !inCriteriaSection && (requirements || !isCheckbox) {
			continue
		}

		var text string
		var completed bool

		if matches := taskCheckboxRegex.FindStringSubmatch(trimmed); len(matches) == 3 {
			mark := strings.ToLower(matches[1])
			completed = !requirements && (mark == "x")
			text = strings.TrimSpace(matches[2])
		} else if matches := listItemRegex.FindStringSubmatch(trimmed); len(matches) == 2 {
			text = strings.TrimSpace(matches[1])
			completed = false
		} else {
			continue
		}

		if text == "" {
			continue
		}

		cleanText := cleanMarkdownFormatting(text)
		if requirements {
			key := strings.ToLower(cleanText)
			if seen[key] {
				continue
			}
			seen[key] = true
		}
		index++
		keywords := extractKeywords(cleanText)

		items = append(items, Criterion{
			ID:        fmt.Sprintf("crit-%d", index),
			Index:     index,
			Raw:       line,
			Text:      cleanText,
			Completed: completed,
			Required:  true,
			Keywords:  keywords,
		})
	}

	return items
}

// ParseAcceptanceCriteria is an alias to ParseChecklist and preserves checked
// states. Use ParseRequirements for requested work that still needs verification.
func ParseAcceptanceCriteria(input string) []Criterion {
	return ParseChecklist(input)
}

// cleanMarkdownFormatting strips bold, italic, code backticks, and trailing punctuation for matching.
func cleanMarkdownFormatting(s string) string {
	s = strings.ReplaceAll(s, "**", "")
	s = strings.ReplaceAll(s, "__", "")
	s = strings.ReplaceAll(s, "`", "")
	s = strings.TrimSpace(s)
	return s
}

// extractKeywords extracts distinct lowercase tokens (length >= 3) for flexible matching.
func extractKeywords(s string) []string {
	words := strings.Fields(strings.ToLower(s))
	var keywords []string
	seen := make(map[string]bool)

	for _, w := range words {
		// Clean punctuation
		cleaned := strings.Trim(w, ":.,;()[]{}*`\"'")
		if len(cleaned) >= 3 && !seen[cleaned] {
			seen[cleaned] = true
			keywords = append(keywords, cleaned)
		}
	}
	return keywords
}
