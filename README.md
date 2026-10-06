<h1 align="center">repolint</h1>

<p align="center">
  Open-source repository linter.<br>
  Does this repo have what a healthy public project needs?
</p>

<p align="center">
  <a href="#quick-start">Quick Start</a> ·
  <a href="#checks">Checks</a> ·
  <a href="https://github.com/Dvorinka/repolint/releases">Releases</a> ·
  <a href="CONTRIBUTING.md">Contributing</a>
</p>

<p align="center">
  <a href="https://github.com/Dvorinka/repolint/actions/workflows/ci.yml"><img src="https://github.com/Dvorinka/repolint/actions/workflows/ci.yml/badge.svg" alt="CI"></a>
  <a href="https://github.com/Dvorinka/repolint/releases"><img src="https://img.shields.io/github/v/release/Dvorinka/repolint" alt="Release"></a>
  <a href="LICENSE"><img src="https://img.shields.io/github/license/Dvorinka/repolint" alt="License"></a>
</p>

## What is repolint?

repolint answers a simple question before you push: **is this repo
actually ready to be public?** LICENSE, README, CONTRIBUTING, SECURITY,
issue/PR templates, dependabot, CI, release automation — and, with
`--remote`, the GitHub-side metadata too: description, topics, detected
license, releases.

Two commands do the whole job: `check` audits, `fix` scaffolds. Missing
files get generated from ecosystem-aware templates — a Go repo gets a Go
CI workflow and a `gomod` dependabot entry, a Node repo gets npm.

> repolint is the successor to the archived `todogroup/repolinter`: same
> job, single Go binary, GitHub-metadata checks, and a `fix` that writes
> the missing files instead of just reporting them.

## Features

- **16 local checks** — license (with SPDX detection), README depth,
  CONTRIBUTING, SECURITY, CoC, CHANGELOG, .gitignore, CI workflow,
  release automation, issue/PR templates, dependabot.
- **Remote metadata checks** — `--remote` adds description, topics,
  GitHub-detected license, and releases via the `gh` CLI. No token
  juggling; it uses your existing `gh` auth.
- **`repolint fix`** — generates missing standard files, tailored to
  the detected ecosystem (Go, Node, Rust, Python). Never overwrites.
  `--dry-run` supported.
- **Agent-ready** — `--json` on every command; stable snake_case keys,
  stable exit codes, `version`, `completion`.

## Quick Start

```bash
# one-liner — latest release binary to ~/.local/bin
curl -fsSL https://raw.githubusercontent.com/Dvorinka/repolint/main/install.sh | sh

# or from source
go install github.com/Dvorinka/repolint/cmd/repolint@latest
```

Then:

```bash
repolint check            # audit current repo — exit 2 if criticals
repolint check --remote   # also check GitHub metadata (needs gh auth)
repolint fix              # scaffold missing files
repolint check --json     # machine-readable for CI/agents
```

## Commands

### `repolint check [--remote] [--json] [--root DIR]`

The gate. Local file checks always run; `--remote` adds GitHub
metadata via `gh api` against the repo's `origin` remote.

Exit codes: `0` clean · `1` warnings only · `2` critical gaps ·
`5` error (unreadable config, malformed input).

### `repolint fix [--dry-run] [--json]`

Creates missing standard files — LICENSE (Apache-2.0), README skeleton,
CONTRIBUTING, SECURITY, issue templates, PR template, dependabot config,
CI workflow — all tailored to the detected ecosystem. Never overwrites
existing files. Remote-metadata gaps are listed as unfixable (they're
`gh repo edit` calls, not files).

### `repolint version` / `repolint completion <bash|zsh|fish>`

Version string and shell completion scripts.

## Checks

| Category | Severity | What it wants |
|---|---|---|
| `license_missing` | critical | LICENSE/COPYING present |
| `readme_missing` | critical | README present |
| `ci_missing` | critical | `.github/workflows/*.yml` or equivalent |
| `license_undetected` | warning | standard SPDX license text |
| `license_mismatch` | warning | matches `expected_license` in config |
| `readme_thin` | warning | install/usage/quickstart section |
| `contributing_missing` | warning | CONTRIBUTING file |
| `security_missing` | warning | SECURITY.md — vuln reporting path |
| `coc_missing` | warning | CODE_OF_CONDUCT |
| `changelog_missing` | warning | CHANGELOG |
| `gitignore_missing` | warning | .gitignore |
| `release_missing` | warning | tag-triggered release automation |
| `issue_tpl_missing` | warning | ISSUE_TEMPLATE forms |
| `pr_tpl_missing` | warning | PULL_REQUEST_TEMPLATE |
| `dependabot_missing` | warning | dependabot or renovate config |
| `desc_missing` | warning | GitHub repo description (`--remote`) |
| `topics_missing` | warning | GitHub topics (`--remote`) |
| `gh_license_missing` | warning | GitHub-detected license (`--remote`) |
| `gh_no_releases` | warning | at least one release (`--remote`) |

`ok` findings are emitted too — `--json` consumers can see full coverage.

## Configuration

`.repolint.yml` — all optional:

```yaml
expected_license: Apache-2.0   # pin the SPDX id
remote: true                   # always run gh-metadata checks
disable: [coc_missing, changelog_missing]
require_files: ["assets/logo.svg"]   # extra must-exist globs

severity:
  changelog_missing: critical  # bump or demote any check
```

## JSON output

```json
{
  "findings": [
    {"severity": "critical", "category": "license_missing",
     "key": "LICENSE", "detail": "no LICENSE/COPYING file — repo isn't legally usable"}
  ],
  "summary": {"critical": 1, "warning": 4, "ok": 11, "total": 16}
}
```

## CI usage

```yaml
- run: repolint check --remote
```

Or use repolint on itself — this repo passes its own lint.

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md).

## Security

See [SECURITY.md](SECURITY.md).

## License

[Apache-2.0](LICENSE)
