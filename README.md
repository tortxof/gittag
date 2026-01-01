# gittag

A simple CLI tool for bumping semantic version tags in Git repositories.

## Usage

```bash
gittag <major|minor|patch>
```

### Examples

```bash
# Current tag: v1.2.3

gittag patch   # Creates v1.2.4
gittag minor   # Creates v1.3.0
gittag major   # Creates v2.0.0
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

## Limitations

This tool currently only works with simple version tags in the format `vMAJOR.MINOR.PATCH` (e.g., `v1.2.3`). Pre-release versions, build metadata, and other extended semver formats are not yet supported.

## Contributing

Pull requests are welcome. This tool is intended to remain simple and should work with version tags that follow [semver](https://semver.org/).

## License

Public domain. See [LICENSE.md](LICENSE.md).
