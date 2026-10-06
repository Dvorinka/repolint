package internal

import "sort"

// Finding is one repo-hygiene result.
type Finding struct {
	Severity string `json:"severity"` // "critical" | "warning" | "ok"
	Category string `json:"category"` // license_missing, ci_missing, ...
	Key      string `json:"key"`      // the thing checked
	Detail   string `json:"detail"`
}

// DefaultSeverity maps category -> default severity.
var DefaultSeverity = map[string]string{
	"license_missing":      "critical",
	"readme_missing":       "critical",
	"ci_missing":           "critical",
	"license_undetected":   "warning",
	"license_mismatch":     "warning",
	"readme_thin":          "warning",
	"contributing_missing": "warning",
	"security_missing":     "warning",
	"coc_missing":          "warning",
	"changelog_missing":    "warning",
	"gitignore_missing":    "warning",
	"release_missing":      "warning",
	"issue_tpl_missing":    "warning",
	"pr_tpl_missing":       "warning",
	"dependabot_missing":   "warning",
	"desc_missing":         "warning",
	"topics_missing":       "warning",
	"gh_license_missing":   "warning",
	"gh_no_releases":       "warning",
	"gh_unreachable":       "warning",
	"missing_file":         "warning",
}

// New builds a finding with the configured severity for its category.
func New(category, key, detail string, overrides map[string]string) Finding {
	sev := DefaultSeverity[category]
	if overrides != nil {
		if s, ok := overrides[category]; ok && s != "" {
			sev = s
		}
	}
	return Finding{Severity: sev, Category: category, Key: key, Detail: detail}
}

// OK is a passing check — kept so --json consumers can see coverage.
func OK(category, key, detail string) Finding {
	return Finding{Severity: "ok", Category: category, Key: key, Detail: detail}
}

// SortFindings orders findings: critical, warning, ok; then category, key.
func SortFindings(fs []Finding) {
	order := map[string]int{"critical": 0, "warning": 1, "ok": 2}
	sort.SliceStable(fs, func(i, j int) bool {
		a, b := fs[i], fs[j]
		if order[a.Severity] != order[b.Severity] {
			return order[a.Severity] < order[b.Severity]
		}
		if a.Category != b.Category {
			return a.Category < b.Category
		}
		return a.Key < b.Key
	})
}

// Summary counts severities.
type Summary struct {
	Critical int `json:"critical"`
	Warning  int `json:"warning"`
	OK       int `json:"ok"`
	Total    int `json:"total"`
}

// Summarize counts.
func Summarize(fs []Finding) Summary {
	var s Summary
	for _, f := range fs {
		switch f.Severity {
		case "critical":
			s.Critical++
		case "warning":
			s.Warning++
		default:
			s.OK++
		}
	}
	s.Total = len(fs)
	return s
}

// Result is the full check output contract.
type Result struct {
	Findings []Finding `json:"findings"`
	Summary  Summary   `json:"summary"`
	Warnings []string  `json:"warnings,omitempty"`
}

// ExitCode maps a result to the repolint contract:
// 0 clean, 1 warnings only, 2 critical findings.
func ExitCode(res Result) int {
	if res.Summary.Critical > 0 {
		return 2
	}
	if res.Summary.Warning > 0 {
		return 1
	}
	return 0
}
