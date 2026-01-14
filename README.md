# 🏷️ gittag

A simple CLI tool for bumping semantic version tags in Git repositories.

## Usage

```bash
gittag <command> [prerelease-id]
```

### Commands

| Command | Description |
|---------|-------------|
| `major` | Bump major version (v1.2.3 → v2.0.0) |
| `minor` | Bump minor version (v1.2.3 → v1.3.0) |
| `patch` | Bump patch version (v1.2.3 → v1.2.4) |
| `pre-major <id>` | Bump major and add pre-release (v1.2.3 → v2.0.0-alpha.1) |
| `pre-minor <id>` | Bump minor and add pre-release (v1.2.3 → v1.3.0-alpha.1) |
| `pre-patch <id>` | Bump patch and add pre-release (v1.2.3 → v1.2.4-alpha.1) |
| `pre <id>` | Update pre-release identifier (v1.3.0-alpha.1 → v1.3.0-alpha.2) |
| `release` | Remove pre-release suffix (v1.3.0-alpha.2 → v1.3.0) |
| `init` | Initialize repo with v0.0.0 tag |

### Flags

| Flag | Description |
|------|-------------|
| `-n`, `--dry-run` | Show what would happen without creating the tag |
| `-v`, `--version` | Print version info |

### Demo

<img src="https://vhs.charm.sh/vhs-1coTf7yL5qFPCR2ayeX6y0.gif" alt="Made with VHS">
<a href="https://vhs.charm.sh">
<img src="https://stuff.charm.sh/vhs/badge.svg">
</a>

### Examples

```bash
# Current tag: v1.2.3

gittag patch   # Creates v1.2.4
gittag minor   # Creates v1.3.0
gittag major   # Creates v2.0.0

# Pre-release workflow
gittag pre-minor alpha.1   # v1.2.3 → v1.3.0-alpha.1
gittag pre alpha.2         # v1.3.0-alpha.1 → v1.3.0-alpha.2
gittag pre beta.1          # v1.3.0-alpha.2 → v1.3.0-beta.1
gittag release             # v1.3.0-beta.1 → v1.3.0
```

## Installation

### mise

```bash
mise use -g github:tortxof/gittag
```

### eget

```bash
eget tortxof/gittag
```

### ubi

```bash
ubi --project tortxof/gittag --in ~/bin
```

### Pre-built binaries

Download from the [releases page](https://github.com/tortxof/gittag/releases).

### Build from source

```bash
go install github.com/tortxof/gittag@latest
```

## How It Works

gittag uses Git commands to discover and create version tags:

1. **Finding the current version**: Runs `git describe --tags --abbrev=0 --match "v*.*.*"` to find the most recent matching tag
2. **Creating new tags**: Runs `git tag <version>` to create lightweight (non-annotated) tags

### Limitations

Because gittag uses `git describe` to find tags, there are some important behaviors to understand:

- **Ancestry-based discovery**: `git describe` only finds tags that are ancestors of the current commit. If you checkout an older commit or a branch that diverged before a tag was created, gittag won't see that tag. It finds the most recent tag *reachable from HEAD*, not the most recent tag in the entire repository.

- **Local tags only**: gittag only sees tags that exist locally. Tags on the remote that haven't been fetched won't be found. Run `git fetch --tags` to sync remote tags before using gittag.

- **Pattern matching**: Tags must match `v*.*.*` to be discovered. Tags like `1.0.0` (no `v` prefix) or `v1.0` (only two components) will be ignored.

- **No tag pushing**: gittag creates tags locally but does not push them. Run `git push --tags` after creating a tag to push it to the remote.

## Notes

- Tags follow the format `vMAJOR.MINOR.PATCH` or `vMAJOR.MINOR.PATCH-PRERELEASE`
- Pre-release identifiers must be valid per [semver spec](https://semver.org/) (dot-separated alphanumeric identifiers)
- You cannot run `major`/`minor`/`patch` on a pre-release version; use `release` first
- Build metadata is parsed but not preserved

## Development

Run tests with `-count=1` to bypass Go's test cache. The integration tests build and run the binary as a subprocess, so Go's cache doesn't detect when `main.go` changes.

```bash
go test -v -count=1 ./...
```

## Contributing

Pull requests are welcome. This tool is intended to remain simple and should work with version tags that follow [semver](https://semver.org/).

## License

Public domain. See [LICENSE.md](LICENSE.md).
