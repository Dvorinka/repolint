// repolint — open-source repository linter.
// Verifies a repo has the files and metadata a healthy public project
// needs: license, docs, CI, release automation, community files, and
// (with --remote) GitHub description/topics/releases.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Dvorinka/repolint/internal"
)

// version is stamped at release: -ldflags "-X main.version=v0.1.0".
var version = "dev"

const usage = `repolint — open-source repository linter

Usage:
  repolint check  [--remote] [--json] [--root DIR] [--config FILE]
  repolint fix    [--dry-run] [--json] [--root DIR] [--config FILE]
  repolint version
  repolint completion <bash|zsh|fish>

Exit codes: 0 clean, 1 warnings only, 2 critical gaps, 5 error.

Local checks need no network. --remote adds GitHub metadata checks
(description, topics, releases) via the gh CLI.
`

func fail(err error) {
	fmt.Fprintf(os.Stderr, "error: %v\n", err)
	os.Exit(5)
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprint(os.Stderr, usage)
		os.Exit(5)
	}
	cmd := os.Args[1]
	switch cmd {
	case "version", "--version", "-version":
		fmt.Println("repolint", version)
		return
	case "completion":
		if len(os.Args) < 3 {
			fmt.Fprint(os.Stderr, usage)
			os.Exit(5)
		}
		s, err := internal.Completion(os.Args[2])
		if err != nil {
			fail(err)
		}
		fmt.Print(s)
		return
	}

	fs := flag.NewFlagSet(cmd, flag.ContinueOnError)
	jsonOut := fs.Bool("json", false, "machine-readable output")
	root := fs.String("root", ".", "repo root to check")
	config := fs.String("config", "", "path to .repolint.yml")
	remote := fs.Bool("remote", false, "include GitHub metadata checks (needs gh)")
	dryRun := fs.Bool("dry-run", false, "show what fix would create")
	if err := fs.Parse(os.Args[2:]); err != nil {
		os.Exit(5)
	}
	abs, err := filepath.Abs(*root)
	if err != nil {
		fail(err)
	}
	cfg, err := internal.LoadConfig(abs, *config)
	if err != nil {
		fail(err)
	}

	switch cmd {
	case "check":
		res, err := internal.RunCheck(abs, cfg, internal.GHCLI{Root: abs}, *remote)
		if err != nil {
			fail(err)
		}
		if *jsonOut {
			internal.WriteJSON(os.Stdout, res)
		} else {
			internal.Report(os.Stdout, res)
		}
		os.Exit(internal.ExitCode(res))
	case "fix":
		res, err := internal.RunCheck(abs, cfg, nil, false)
		if err != nil {
			fail(err)
		}
		fr, err := internal.Fix(abs, res, *dryRun)
		if err != nil {
			fail(err)
		}
		if *jsonOut {
			b, _ := json.MarshalIndent(fr, "", "  ")
			fmt.Println(string(b))
			return
		}
		verb := "Created"
		if *dryRun {
			verb = "Would create"
		}
		for _, p := range fr.Created {
			fmt.Printf("  %s %s\n", verb, p)
		}
		for _, p := range fr.Skipped {
			fmt.Printf("  exists     %s\n", p)
		}
		if len(fr.Unfixable) > 0 {
			fmt.Printf("%d finding(s) have no scaffold (remote metadata, README content) — fix by hand.\n",
				len(fr.Unfixable))
		}
	default:
		fmt.Fprintf(os.Stderr, "unknown command %q\n\n%s", cmd, usage)
		os.Exit(5)
	}
}
