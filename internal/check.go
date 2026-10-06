package internal

// RunCheck executes all enabled checks: local file rules plus remote
// GitHub metadata when cfg.Remote or the flag is set.
func RunCheck(root string, cfg Config, gh GH, remote bool) (Result, error) {
	findings := LocalChecks(root, cfg)
	var warnings []string
	if remote || cfg.Remote {
		if gh == nil {
			warnings = append(warnings, "no gh interface — remote checks skipped")
		} else {
			remoteFindings, _ := RemoteChecks(gh, cfg)
			findings = append(findings, remoteFindings...)
		}
	}
	SortFindings(findings)
	return Result{Findings: findings, Summary: Summarize(findings), Warnings: warnings}, nil
}
