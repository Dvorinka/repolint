package internal_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Dvorinka/repolint/internal"
)

// bareRepo returns a temp dir that fails almost everything.
func bareRepo(t *testing.T) string {
	t.Helper()
	return t.TempDir()
}

// fullRepo returns a temp dir that should pass all local checks.
func fullRepo(t *testing.T) string {
	t.Helper()
	r := t.TempDir()
	write := func(rel, body string) {
		p := filepath.Join(r, rel)
		os.MkdirAll(filepath.Dir(p), 0o755)
		os.WriteFile(p, []byte(body), 0o644)
	}
	write("LICENSE", "Apache License\n   Version 2.0, January 2004\n")
	write("README.md", "# x\n\n## Install\n\n```\ninstall it\n```\n")
	write("CONTRIBUTING.md", "c")
	write("SECURITY.md", "s")
	write("CODE_OF_CONDUCT.md", "c")
	write("CHANGELOG.md", "c")
	write(".gitignore", ".env\n")
	write(".github/workflows/ci.yml", "on: [push]\njobs:\n  t:\n    steps: []")
	write(".github/workflows/release.yml", "on:\n  push:\n    tags: [v*]\njobs:\n  r:\n    steps: []")
	write(".github/ISSUE_TEMPLATE/bug.yml", "name: bug")
	write(".github/PULL_REQUEST_TEMPLATE.md", "## what")
	write(".github/dependabot.yml", "version: 2")
	return r
}

func cats(res internal.Result) map[string]string {
	m := map[string]string{}
	for _, f := range res.Findings {
		m[f.Category] = f.Severity
	}
	return m
}

func TestBareRepoFlagsCriticals(t *testing.T) {
	cfg, _ := internal.LoadConfig(bareRepo(t), "")
	res, err := internal.RunCheck(bareRepo(t), cfg, nil, false)
	if err != nil {
		t.Fatal(err)
	}
	c := cats(res)
	for _, cat := range []string{"license_missing", "readme_missing", "ci_missing"} {
		if c[cat] != "critical" {
			t.Fatalf("%s = %q, want critical", cat, c[cat])
		}
	}
	if res.Summary.Critical < 3 {
		t.Fatalf("summary %+v", res.Summary)
	}
	if internal.ExitCode(res) != 2 {
		t.Fatal("exit should be 2")
	}
}

func TestFullRepoClean(t *testing.T) {
	root := fullRepo(t)
	cfg, _ := internal.LoadConfig(root, "")
	res, err := internal.RunCheck(root, cfg, nil, false)
	if err != nil {
		t.Fatal(err)
	}
	if res.Summary.Critical != 0 || res.Summary.Warning != 0 {
		t.Fatalf("full repo should be clean: %+v", res.Summary)
	}
	if internal.ExitCode(res) != 0 {
		t.Fatal("exit should be 0")
	}
}

func TestLicenseDetection(t *testing.T) {
	r := bareRepo(t)
	os.WriteFile(filepath.Join(r, "LICENSE"), []byte("MIT License\n\nPermission is hereby granted"), 0o644)
	cfg, _ := internal.LoadConfig(r, "")
	cfg.ExpectedLicense = "MIT"
	res, _ := internal.RunCheck(r, cfg, nil, false)
	c := cats(res)
	if c["license_undetected"] != "ok" {
		t.Fatalf("MIT license not detected: %v", res.Findings)
	}
	// wrong expectation fires
	cfg.ExpectedLicense = "Apache-2.0"
	res, _ = internal.RunCheck(r, cfg, nil, false)
	if cats(res)["license_mismatch"] != "warning" {
		t.Fatal("mismatched expected_license should warn")
	}
}

func TestDisableAndRequireFiles(t *testing.T) {
	r := bareRepo(t)
	os.WriteFile(filepath.Join(r, ".repolint.yml"), []byte(
		"disable: [coc_missing, changelog_missing]\nrequire_files: [must_exist.txt]\n"), 0o644)
	cfg, err := internal.LoadConfig(r, "")
	if err != nil {
		t.Fatal(err)
	}
	res, _ := internal.RunCheck(r, cfg, nil, false)
	c := cats(res)
	if _, ok := c["coc_missing"]; ok {
		t.Fatal("disabled check still ran")
	}
	if c["missing_file"] != "warning" {
		t.Fatal("require_files gap not flagged")
	}
}

// fakeGH implements internal.GH for remote-check tests.
type fakeGH struct {
	repo   string
	bodies map[string]string
	errs   map[string]error
}

func (f fakeGH) Repo() string { return f.repo }
func (f fakeGH) API(ep string) (string, error) {
	if err := f.errs[ep]; err != nil {
		return "", err
	}
	return f.bodies[ep], nil
}

func TestRemoteChecks(t *testing.T) {
	gh := fakeGH{
		repo: "me/x",
		bodies: map[string]string{
			"repos/me/x":                     `{"description":"","topics":[],"default_branch":"main","license":null}`,
			"repos/me/x/releases?per_page=1": `[{"tag_name":"v1"}]`,
		},
	}
	fs, repo := internal.RemoteChecks(gh, internal.DefaultConfig())
	if repo != "me/x" {
		t.Fatal("repo not resolved")
	}
	m := map[string]string{}
	for _, f := range fs {
		m[f.Category] = f.Severity
	}
	if m["desc_missing"] != "warning" || m["topics_missing"] != "warning" ||
		m["gh_license_missing"] != "warning" {
		t.Fatalf("expected 3 warnings: %v", fs)
	}
	if m["gh_no_releases"] != "ok" {
		t.Fatal("has release should be ok")
	}
}

func TestFixScaffolds(t *testing.T) {
	r := bareRepo(t)
	os.WriteFile(filepath.Join(r, "go.mod"), []byte("module x"), 0o644)
	cfg, _ := internal.LoadConfig(r, "")
	res, _ := internal.RunCheck(r, cfg, nil, false)
	fr, err := internal.Fix(r, res, true)
	if err != nil {
		t.Fatal(err)
	}
	if len(fr.Created) == 0 {
		t.Fatal("dry-run should list creations")
	}
	for _, p := range fr.Created {
		if _, err := os.Stat(filepath.Join(r, p)); !os.IsNotExist(err) {
			t.Fatalf("dry-run wrote %s", p)
		}
	}
	fr, _ = internal.Fix(r, res, false)
	if len(fr.Created) == 0 {
		t.Fatal("fix created nothing")
	}
	// go ecosystem → gomod dependabot
	b, _ := os.ReadFile(filepath.Join(r, ".github/dependabot.yml"))
	if got := string(b); !containsStr(got, "gomod") {
		t.Fatalf("dependabot missing gomod: %s", got)
	}
	// re-check: criticals should be gone
	res2, _ := internal.RunCheck(r, cfg, nil, false)
	if res2.Summary.Critical != 0 {
		t.Fatalf("fix left criticals: %+v", res2.Summary)
	}
	// idempotent — second fix creates nothing
	fr2, _ := internal.Fix(r, res2, false)
	if len(fr2.Created) != 0 {
		t.Fatalf("fix not idempotent: %v", fr2.Created)
	}
}

func containsStr(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
