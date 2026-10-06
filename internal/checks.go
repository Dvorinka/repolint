package internal

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Check groups: local file checks run always; remote checks (--remote)
// use `gh api` against the repo's origin remote.

func exists(root string, globs ...string) string {
	for _, g := range globs {
		m, _ := filepath.Glob(filepath.Join(root, g))
		if len(m) > 0 {
			return m[0]
		}
	}
	return ""
}

func readSmall(path string, max int64) string {
	f, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer f.Close()
	b := make([]byte, max)
	n, _ := f.Read(b)
	return string(b[:n])
}

// fileCheck emits one finding per check: ok or missing.
type fileCheck struct {
	cat     string
	key     string
	okMsg   string
	missMsg string
	run     func(root string) bool
}

var licenseRe = map[string]*regexp.Regexp{
	"Apache-2.0": regexp.MustCompile(`Apache License`),
	"MIT":        regexp.MustCompile(`MIT License|Permission is hereby granted`),
	"GPL-3.0":    regexp.MustCompile(`GNU GENERAL PUBLIC LICENSE`),
	"BSD-3":      regexp.MustCompile(`Redistribution and use in source and binary`),
}

func detectLicense(root string) (file, spdx string) {
	f := exists(root, "LICENSE*", "LICENCE*", "COPYING*", "license*")
	if f == "" {
		return "", ""
	}
	head := readSmall(f, 4096)
	for spdx, re := range licenseRe {
		if re.MatchString(head) {
			return f, spdx
		}
	}
	return f, "NOASSERTION"
}

func licenseMsg(lfile, spdx string) string {
	if lfile == "" {
		return "n/a — no license file"
	}
	return "detected " + spdx
}

func localChecks(root string) []fileCheck {
	lfile, spdx := detectLicense(root)
	readmeOK := "README has install/usage content"
	if exists(root, "README*", "readme*") == "" {
		readmeOK = "n/a — no README"
	}
	return []fileCheck{
		{cat: "license_missing", key: "LICENSE",
			okMsg: "found", missMsg: "no LICENSE/COPYING file — repo isn't legally usable",
			run: func(string) bool { return lfile != "" }},
		{cat: "license_undetected", key: "license_spdx",
			okMsg: licenseMsg(lfile, spdx), missMsg: "license text not recognized as a standard SPDX license",
			run: func(string) bool { return lfile == "" || (spdx != "" && spdx != "NOASSERTION") }},
		{cat: "readme_missing", key: "README",
			okMsg: "found", missMsg: "no README — first thing every visitor sees",
			run: func(r string) bool { return exists(r, "README*", "readme*") != "" }},
		{cat: "readme_thin", key: "readme_content",
			okMsg:   readmeOK,
			missMsg: "README has no install or usage section",
			run: func(r string) bool {
				p := exists(r, "README*", "readme*")
				if p == "" {
					return true // readme_missing covers it
				}
				body := strings.ToLower(readSmall(p, 1<<20))
				return strings.Contains(body, "install") ||
					strings.Contains(body, "quick start") ||
					strings.Contains(body, "quickstart") ||
					strings.Contains(body, "usage")
			}},
		{cat: "contributing_missing", key: "CONTRIBUTING",
			okMsg: "found", missMsg: "no CONTRIBUTING file",
			run: func(r string) bool { return exists(r, "CONTRIBUTING*") != "" }},
		{cat: "security_missing", key: "SECURITY",
			okMsg: "found", missMsg: "no SECURITY.md — nowhere to report vulns",
			run: func(r string) bool {
				return exists(r, "SECURITY*", ".github/SECURITY*") != ""
			}},
		{cat: "coc_missing", key: "CODE_OF_CONDUCT",
			okMsg: "found", missMsg: "no code of conduct",
			run: func(r string) bool {
				return exists(r, "CODE_OF_CONDUCT*", ".github/CODE_OF_CONDUCT*") != ""
			}},
		{cat: "changelog_missing", key: "CHANGELOG",
			okMsg: "found", missMsg: "no CHANGELOG — releases have no written history",
			run: func(r string) bool { return exists(r, "CHANGELOG*") != "" }},
		{cat: "gitignore_missing", key: ".gitignore",
			okMsg: "found", missMsg: "no .gitignore",
			run: func(r string) bool {
				_, err := os.Stat(filepath.Join(r, ".gitignore"))
				return err == nil
			}},
		{cat: "ci_missing", key: "ci_workflow",
			okMsg: "workflow found", missMsg: "no CI workflow (.github/workflows/*.yml)",
			run: func(r string) bool {
				return exists(r, ".github/workflows/*.yml", ".github/workflows/*.yaml",
					".gitlab-ci.yml", "Jenkinsfile", ".circleci/config.yml") != ""
			}},
		{cat: "release_missing", key: "release_workflow",
			okMsg:   "release workflow found",
			missMsg: "no release automation (tag-triggered workflow, goreleaser, etc.)",
			run: func(r string) bool {
				for _, g := range []string{".github/workflows/*.yml", ".github/workflows/*.yaml"} {
					files, _ := filepath.Glob(filepath.Join(r, g))
					for _, f := range files {
						body := readSmall(f, 1<<20)
						if strings.Contains(body, "tags:") ||
							strings.Contains(body, "goreleaser") ||
							strings.Contains(strings.ToLower(body), "release") {
							return true
						}
					}
				}
				return exists(r, ".goreleaser*", "goreleaser*") != ""
			}},
		{cat: "issue_tpl_missing", key: "issue_templates",
			okMsg: "found", missMsg: "no issue templates",
			run: func(r string) bool {
				return exists(r, ".github/ISSUE_TEMPLATE/*", ".github/ISSUE_TEMPLATE.md") != ""
			}},
		{cat: "pr_tpl_missing", key: "pr_template",
			okMsg: "found", missMsg: "no pull request template",
			run: func(r string) bool {
				return exists(r, ".github/PULL_REQUEST_TEMPLATE*",
					".github/pull_request_template*", "PULL_REQUEST_TEMPLATE*") != ""
			}},
		{cat: "dependabot_missing", key: "dependabot",
			okMsg: "found", missMsg: "no dependabot config — deps never get update PRs",
			run: func(r string) bool {
				return exists(r, ".github/dependabot*", ".github/renovate*",
					"renovate.json") != ""
			}},
	}
}

// LocalChecks runs the local file rule set.
func LocalChecks(root string, cfg Config) []Finding {
	var out []Finding
	for _, c := range localChecks(root) {
		if cfg.Disabled[c.cat] {
			continue
		}
		if c.run(root) {
			out = append(out, OK(c.cat, c.key, c.okMsg))
		} else {
			out = append(out, New(c.cat, c.key, c.missMsg, cfg.Severity))
		}
	}
	// expected_license: pin the SPDX id when the project standardizes on one
	if cfg.ExpectedLicense != "" {
		_, spdx := detectLicense(root)
		if spdx == cfg.ExpectedLicense {
			out = append(out, OK("license_mismatch", "expected_license",
				"license is "+spdx))
		} else {
			out = append(out, New("license_mismatch", "expected_license",
				"expected "+cfg.ExpectedLicense+", detected "+orNone(spdx), cfg.Severity))
		}
	}
	// require_files: user-declared must-exist globs
	for _, g := range cfg.ExtraFiles {
		if exists(root, g) != "" {
			out = append(out, OK("missing_file", g, "found"))
		} else {
			out = append(out, New("missing_file", g,
				"required by .repolint.yml but not found", cfg.Severity))
		}
	}
	return out
}
