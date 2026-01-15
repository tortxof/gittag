# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

gittag is a simple Go CLI tool for bumping semantic version tags in Git repositories. It reads the current git tag matching `v*.*.*`, bumps it according to the specified level (major/minor/patch), and creates a new tag.

## Commands

### Build
```bash
go build
```

### Run tests

The integration tests have an issue with the Go build cache. Use the `-count=1` option.

```bash
go test -v -count=1 ./...
```

### Install locally
```bash
go install
```

### Release
Releases are handled by GoReleaser. Tests run automatically before release via the goreleaser hook.

## Architecture

This is a single-file Go application (`main.go`) with accompanying tests (`main_test.go`).

**Key types and functions:**
- `Version` struct - holds Major, Minor, Patch integers
- `Version.Bump(level)` - returns new Version with appropriate field incremented
- `Version.String()` - formats as `vX.Y.Z`
- `ParseVersion(tag)` - parses semver string (with optional prerelease/metadata) into Version
- `GetCurrentTag()` - runs `git describe` to get the latest matching tag
- `AddVersionTag(v)` - runs `git tag` to create the new tag

**CLI flags:**
- `-v`, `-version`: print version info
- `-n`, `-dry-run`: show what would happen without creating the tag

## Version Format

Tags must follow `vMAJOR.MINOR.PATCH` format. The parser accepts optional prerelease and build metadata suffixes per semver spec, but only the major/minor/patch components are used.
