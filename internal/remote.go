package internal

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
)

// GH abstracts the gh CLI so tests can fake it.
type GH interface {
	// API runs `gh api <endpoint>` and returns stdout.
	API(endpoint string) (string, error)
	// Repo returns "owner/repo" for the current dir's origin remote, "" if none.
	Repo() string
}

// GHCLI is the real gh-backed implementation.
type GHCLI struct{ Root string }

func (g GHCLI) API(endpoint string) (string, error) {
	out, err := exec.Command("gh", "api", endpoint).Output()
	if err != nil {
		return "", fmt.Errorf("gh api %s: %w", endpoint, err)
	}
	return string(out), nil
}

var repoRe = regexp.MustCompile(`github\.com[:/]+([A-Za-z0-9_.-]+/[A-Za-z0-9_.-]+?)(\.git)?$`)

func (g GHCLI) Repo() string {
	out, err := exec.Command("git", "-C", g.Root, "remote", "get-url", "origin").Output()
	if err != nil {
		return ""
	}
	m := repoRe.FindStringSubmatch(strings.TrimSpace(string(out)))
	if len(m) < 2 {
		return ""
	}
	return m[1]
}

// RemoteChecks queries GitHub repo metadata: description, topics,
// detected license, default branch, release count. Everything the local
// filesystem can't see.
func RemoteChecks(g GH, cfg Config) ([]Finding, string) {
	var out []Finding
	repo := g.Repo()
	if repo == "" {
		return []Finding{New("gh_unreachable", "origin",
			"no GitHub origin remote — remote checks skipped", cfg.Severity)}, ""
	}
	body, err := g.API("repos/" + repo)
	if err != nil {
		return []Finding{New("gh_unreachable", repo,
			"gh api failed (auth? network?): "+err.Error(), cfg.Severity)}, repo
	}
	var meta struct {
		Description   string   `json:"description"`
		Topics        []string `json:"topics"`
		DefaultBranch string   `json:"default_branch"`
		License       *struct {
			SPDX string `json:"spdx_id"`
		} `json:"license"`
	}
	if err := json.Unmarshal([]byte(body), &meta); err != nil {
		return []Finding{New("gh_unreachable", repo, "unparseable repo metadata", cfg.Severity)}, repo
	}
	add := func(cat, key, okMsg, missMsg string, pass bool) {
		if cfg.Disabled[cat] {
			return
		}
		if pass {
			out = append(out, OK(cat, key, okMsg))
		} else {
			out = append(out, New(cat, key, missMsg, cfg.Severity))
		}
	}
	add("desc_missing", repo, "description set",
		"repo has no description — shows as blank on GitHub",
		strings.TrimSpace(meta.Description) != "")
	add("topics_missing", repo, fmt.Sprintf("%d topics", len(meta.Topics)),
		"no topics — invisible to GitHub search/topic pages",
		len(meta.Topics) > 0)
	add("gh_license_missing", repo, "GitHub detected license",
		"GitHub shows 'View license' — LICENSE not machine-detectable",
		meta.License != nil && meta.License.SPDX != "" && meta.License.SPDX != "NOASSERTION")

	rel, relErr := g.API("repos/" + repo + "/releases?per_page=1")
	add("gh_no_releases", repo, "has releases",
		"no releases — users have no binaries or tagged versions",
		relErr == nil && strings.Contains(rel, `"tag_name"`))
	return out, repo
}

func orNone(s string) string {
	if s == "" {
		return "none"
	}
	return s
}
