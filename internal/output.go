package internal

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

// WriteJSON emits the machine-readable contract.
func WriteJSON(w io.Writer, res Result) {
	if res.Findings == nil {
		res.Findings = []Finding{}
	}
	b, _ := json.MarshalIndent(res, "", "  ")
	fmt.Fprintln(w, string(b))
}

var sevGlyph = map[string]string{
	"critical": "✗",
	"warning":  "△",
	"ok":       "✓",
}

// Report prints human-readable findings grouped by severity.
func Report(w io.Writer, res Result) {
	fmt.Fprintf(w, "Repository lint — %s\n\n", summaryLine(res.Summary))
	last := ""
	for _, f := range res.Findings {
		if f.Severity != last {
			fmt.Fprintf(w, "\n%s:\n", strings.ToUpper(f.Severity))
			last = f.Severity
		}
		fmt.Fprintf(w, "  %s %-26s %s\n", sevGlyph[f.Severity], f.Category, f.Detail)
	}
	for _, warn := range res.Warnings {
		fmt.Fprintf(w, "\nnote: %s\n", warn)
	}
}

func summaryLine(s Summary) string {
	if s.Critical == 0 && s.Warning == 0 {
		return fmt.Sprintf("clean (%d checks pass)", s.OK)
	}
	var parts []string
	if s.Critical > 0 {
		parts = append(parts, fmt.Sprintf("%d critical", s.Critical))
	}
	if s.Warning > 0 {
		parts = append(parts, fmt.Sprintf("%d warnings", s.Warning))
	}
	return fmt.Sprintf("%s, %d checks pass", strings.Join(parts, ", "), s.OK)
}
