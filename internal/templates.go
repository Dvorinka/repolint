package internal

import "fmt"

// Template bodies for the fix scaffolder. Keep them minimal and honest —
// a starting point, not a fake finished file.

func readme(name string) string {
	return fmt.Sprintf(`# %s

<!-- one line: what it does and who it's for -->

## Install

<!-- how to get it -->

## Usage

<!-- the 30-second example -->

## License

[Apache-2.0](LICENSE)
`, name)
}

func contributing(name string) string {
	return fmt.Sprintf(`# Contributing to %s

## Setup

Clone, install dependencies, run the tests.

## Workflow

1. Branch from main.
2. Make the change; add a test for non-trivial logic.
3. Run build + tests + lint locally.
4. Open a PR describing what and why.

## Style

Match the surrounding code. No new dependencies without a reason.
`, name)
}

func security(name string) string {
	return fmt.Sprintf(`# Security Policy

## Reporting a vulnerability

Do **not** open a public issue. Report privately to the maintainers
(see %s owner profile for contact) with a description and reproduction.

You will get an acknowledgement within a few days. Please allow
reasonable time for a fix before public disclosure.
`, name)
}

func coc(name string) string {
	return fmt.Sprintf(`# Code of Conduct

Be respectful. No harassment, no personal attacks, no spam.
Maintainers may remove contributions and participants that violate this.

For %s specifics, contact the maintainers directly.
`, name)
}

func gitignore(eco string) string {
	base := ".env\n*.local\n.DS_Store\n"
	switch eco {
	case "go":
		return base + "*.out\n*.test\n/bin/\n"
	case "node":
		return base + "node_modules/\ndist/\ncoverage/\n"
	case "rust":
		return base + "target/\n"
	case "python":
		return base + "__pycache__/\n*.pyc\n.venv/\ndist/\n"
	}
	return base
}

func dependabot(eco string) string {
	var pkgs string
	switch eco {
	case "go":
		pkgs = "  - package-ecosystem: gomod\n    directory: /\n    schedule: { interval: weekly }\n"
	case "node":
		pkgs = "  - package-ecosystem: npm\n    directory: /\n    schedule: { interval: weekly }\n"
	case "rust":
		pkgs = "  - package-ecosystem: cargo\n    directory: /\n    schedule: { interval: weekly }\n"
	case "python":
		pkgs = "  - package-ecosystem: pip\n    directory: /\n    schedule: { interval: weekly }\n"
	}
	return "version: 2\nupdates:\n" + pkgs +
		"  - package-ecosystem: github-actions\n    directory: /\n    schedule: { interval: weekly }\n"
}

func ciWorkflow(eco, name string) string {
	switch eco {
	case "go":
		return fmt.Sprintf(`name: ci

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version-file: go.mod
      - run: go vet ./...
      - run: go test ./...
      - run: go build ./...
`)
	case "node":
		return `name: ci

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
        with:
          node-version: 22
      - run: npm ci
      - run: npm test
`
	}
	return `name: ci

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      # add your build/test steps
`
}

func releaseWorkflow(eco, name string) string {
	if eco == "go" {
		return fmt.Sprintf(`name: release

on:
  push:
    tags: ["v*"]

permissions:
  contents: write

jobs:
  release:
    runs-on: ubuntu-latest
    strategy:
      matrix:
        include:
          - { goos: linux, goarch: amd64 }
          - { goos: linux, goarch: arm64 }
          - { goos: darwin, goarch: amd64 }
          - { goos: darwin, goarch: arm64 }
          - { goos: windows, goarch: amd64 }
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version-file: go.mod
      - name: build
        run: |
          ver=${GITHUB_REF_NAME}
          out=%s-${ver}-${{ matrix.goos }}-${{ matrix.goarch }}
          [ "${{ matrix.goos }}" = windows ] && out=$out.exe
          CGO_ENABLED=0 GOOS=${{ matrix.goos }} GOARCH=${{ matrix.goarch }} \
            go build -ldflags "-s -w -X main.version=$ver" -o "$out" ./cmd/%s
      - uses: softprops/action-gh-release@v2
        with:
          files: %s-*
`, name, name, name)
	}
	return `name: release

on:
  push:
    tags: ["v*"]

permissions:
  contents: write

jobs:
  release:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: softprops/action-gh-release@v2
        # add build steps + files: before this
`
}

func prTemplate() string {
	return `## What

<!-- one paragraph or a few bullets -->

## Verification

- [ ] builds clean
- [ ] tests pass
- [ ] lint clean
- [ ] smoke-tested
`
}

func bugTemplate() string {
	return `name: Bug report
description: Something doesn't work as documented
labels: ["bug"]
body:
  - type: textarea
    id: what
    attributes:
      label: What happened
      description: Command run, expected output, actual output.
    validations:
      required: true
  - type: textarea
    id: env
    attributes:
      label: Environment
      description: OS, version, how installed.
  - type: textarea
    id: repro
    attributes:
      label: Minimal reproduction
`
}

func featureTemplate() string {
	return `name: Feature request
description: A capability that would make the project more useful
labels: ["enhancement"]
body:
  - type: textarea
    id: problem
    attributes:
      label: Problem
      description: What are you trying to do that the project can't?
    validations:
      required: true
  - type: textarea
    id: proposal
    attributes:
      label: Proposed behavior
`
}
