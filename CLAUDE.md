# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

gittag is a Go CLI tool for bumping semantic version tags in Git repositories. It reads the current version (from git tags or a VERSION file), bumps it according to the specified level, and creates a new tag. Supports prerelease versions and can work from any subdirectory within a git repository.

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

The application is split into multiple Go files:

- `main.go` - CLI entry point, flag parsing, usage text
- `version.go` - Version struct, parsing, file I/O
- `git.go` - Git operations (tags, commits, staging)
- `operations.go` - Operation handlers (DoBump, DoInit, etc.)
- `main_test.go` - Unit tests
- `integration_test.go` - Integration tests (builds binary and tests CLI)

**Key types:**
- `Version` struct - holds Major, Minor, Patch integers and Prerelease string
- `Options` struct - holds CLI options (OpMode, PrereleaseID, UseFile, VersionFile, DryRun)
- `Operation` func type - `func(Options) error` signature for operation handlers
- `operations` map - maps command names to Operation handlers

**Version functions (version.go):**
- `ParseVersion(tag)` - parses version string into Version struct
- `GetVersionFromFile(path)` - reads version string from file
- `WriteVersionToFile(path, v)` - writes version to file (without 'v' prefix)

**Version methods:**
- `Bump(level)` - returns new Version with appropriate field incremented
- `String()` - formats as `vX.Y.Z` or `vX.Y.Z-prerelease`
- `SetPrerelease(id)` - returns new Version with prerelease identifier (returns error if invalid)
- `ClearPrerelease()` - returns Version without prerelease
- `IsPrerelease()` - checks if version has prerelease
- `ShouldCreateTag()` - true for releases and numbered prereleases (e.g., alpha.1)

**Git functions (git.go):**
- `GetCurrentTag()` - runs `git describe` to get the latest matching tag
- `AddVersionTag(v)` - creates a git tag
- `HasTag()` - checks if any version tag exists
- `HasStagedChanges()` - checks for staged git changes
- `HasFileChanges(path)` - checks for uncommitted changes to a file
- `StageFile(path)` - stages a file
- `CreateCommit(message)` - creates a git commit
- `GetRepoRoot()` - gets git repository root directory

**Operations (operations.go):**
- `DoBump` - bump major/minor/patch (blocks if current is prerelease)
- `DoBumpPre` - bump and add prerelease identifier
- `DoPre` - update prerelease identifier on existing prerelease
- `DoRelease` - remove prerelease to create release version
- `DoInit` - initialize with v0.0.0 tag

**Helper functions (operations.go):**
- `GetCurrentVersion(o)` - gets current version from file or git tag
- `ApplyVersion(currentVersion, nextVersion, o)` - applies the new version (writes file and/or creates tag)
- `ValidateGitStateForFileMode(versionFile)` - checks for staged changes and uncommitted file changes
- `CommitAndTag(versionFile, v)` - stages the version file, commits it, and creates a tag

**CLI commands:**
- `init` - Initialize with v0.0.0 tag (only if no version tag exists)
- `major`, `minor`, `patch` - Bump version
- `pre-major`, `pre-minor`, `pre-patch` - Bump and add prerelease
- `pre <id>` - Update prerelease identifier
- `release` - Remove prerelease to create release

**CLI flags:**
- `-v`, `-version`: print version info
- `-n`, `-dry-run`: show what would happen without creating the tag
- `-f`: read version from file 'VERSION' (in repo root)
- `-file <path>`: read version from specified file
- `-bash-completion`: output bash completion script

## Version Format

Tags must follow `vMAJOR.MINOR.PATCH` or `vMAJOR.MINOR.PATCH-prerelease` format. The parser accepts optional build metadata suffixes per semver spec, but only the major/minor/patch/prerelease components are used.

## File Mode

When using `-f` or `-file`, the VERSION file is updated and committed before creating the git tag (ensuring the tag points to the correct version). Git tags are only created for release versions and numbered prereleases (e.g., alpha.1, rc1). Unnumbered prereleases (e.g., alpha, beta) only update the file.
