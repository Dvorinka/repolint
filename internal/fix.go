package internal

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// FixResult describes what fix did or would do.
type FixResult struct {
	Created   []string `json:"created"`
	Skipped   []string `json:"skipped"`   // already existed
	Unfixable []string `json:"unfixable"` // findings no template covers
}

// FixFiles is what `fix` can generate. Content comes from Templates.
var fixOrder = []string{
	"README.md", "LICENSE", "CONTRIBUTING.md", "SECURITY.md", "CODE_OF_CONDUCT.md",
	"CHANGELOG.md", ".gitignore",
	".github/ISSUE_TEMPLATE/bug_report.yml",
	".github/ISSUE_TEMPLATE/feature_request.yml",
	".github/ISSUE_TEMPLATE/config.yml",
	".github/PULL_REQUEST_TEMPLATE.md",
	".github/dependabot.yml",
	".github/workflows/ci.yml",
}

// catToFiles maps a failing category to the files fix generates for it.
var catToFiles = map[string][]string{
	"license_missing":      {"LICENSE"},
	"license_undetected":   {"LICENSE"},
	"readme_missing":       {"README.md"},
	"contributing_missing": {"CONTRIBUTING.md"},
	"security_missing":     {"SECURITY.md"},
	"coc_missing":          {"CODE_OF_CONDUCT.md"},
	"changelog_missing":    {"CHANGELOG.md"},
	"gitignore_missing":    {".gitignore"},
	"ci_missing":           {".github/workflows/ci.yml"},
	"release_missing":      {".github/workflows/release.yml"},
	"issue_tpl_missing": {
		".github/ISSUE_TEMPLATE/bug_report.yml",
		".github/ISSUE_TEMPLATE/feature_request.yml",
		".github/ISSUE_TEMPLATE/config.yml",
	},
	"pr_tpl_missing":     {".github/PULL_REQUEST_TEMPLATE.md"},
	"dependabot_missing": {".github/dependabot.yml"},
}

// Fix creates missing standard files. Never overwrites. dryRun lists
// what would be created without writing. Ecosystem-aware templates use
// Ecosystem(root) to pick CI/dependabot content.
func Fix(root string, res Result, dryRun bool) (FixResult, error) {
	fr := FixResult{}
	want := map[string]bool{}
	fixableCats := map[string]bool{}
	for _, f := range res.Findings {
		if f.Severity == "ok" {
			continue
		}
		files, ok := catToFiles[f.Category]
		if !ok {
			fr.Unfixable = append(fr.Unfixable, f.Category)
			continue
		}
		fixableCats[f.Category] = true
		for _, p := range files {
			want[p] = true
		}
	}
	// Deterministic order from fixOrder + release.yml.
	order := append(append([]string{}, fixOrder...), ".github/workflows/release.yml")
	for _, rel := range order {
		if !want[rel] {
			continue
		}
		full := filepath.Join(root, rel)
		if _, err := os.Stat(full); err == nil {
			fr.Skipped = append(fr.Skipped, rel)
			continue
		}
		if dryRun {
			fr.Created = append(fr.Created, rel)
			continue
		}
		body, err := Template(rel, root)
		if err != nil {
			return fr, err
		}
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			return fr, err
		}
		if err := os.WriteFile(full, []byte(body), 0o644); err != nil {
			return fr, err
		}
		fr.Created = append(fr.Created, rel)
	}
	return fr, nil
}

// Ecosystem sniffs the repo's primary stack for template content.
func Ecosystem(root string) string {
	for _, eco := range []struct {
		name string
		file string
	}{
		{"go", "go.mod"},
		{"node", "package.json"},
		{"rust", "Cargo.toml"},
		{"python", "pyproject.toml"},
	} {
		if _, err := os.Stat(filepath.Join(root, eco.file)); err == nil {
			return eco.name
		}
	}
	return "generic"
}

// Template returns scaffold content for a path, tailored to the repo.
func Template(rel, root string) (string, error) {
	eco := Ecosystem(root)
	name := filepath.Base(root)
	switch rel {
	case "README.md":
		return readme(name), nil
	case "LICENSE":
		return apache20(), nil
	case "CONTRIBUTING.md":
		return contributing(name), nil
	case "SECURITY.md":
		return security(name), nil
	case "CODE_OF_CONDUCT.md":
		return coc(name), nil
	case "CHANGELOG.md":
		return "# Changelog\n\nAll notable changes to this project are documented here.\n\nThe format follows [Keep a Changelog](https://keepachangelog.com/).\n\n## [Unreleased]\n", nil
	case ".gitignore":
		return gitignore(eco), nil
	case ".github/dependabot.yml":
		return dependabot(eco), nil
	case ".github/workflows/ci.yml":
		return ciWorkflow(eco, name), nil
	case ".github/workflows/release.yml":
		return releaseWorkflow(eco, name), nil
	case ".github/PULL_REQUEST_TEMPLATE.md":
		return prTemplate(), nil
	case ".github/ISSUE_TEMPLATE/bug_report.yml":
		return bugTemplate(), nil
	case ".github/ISSUE_TEMPLATE/feature_request.yml":
		return featureTemplate(), nil
	case ".github/ISSUE_TEMPLATE/config.yml":
		return "blank_issues_enabled: true\n", nil
	}
	return "", fmt.Errorf("no template for %s", rel)
}

func apache20() string {
	return fmt.Sprintf(`Apache License
                           Version 2.0, January 2004
                        http://www.apache.org/licenses/

   Copyright %d

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
`, time.Now().Year())
}
