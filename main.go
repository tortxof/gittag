package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

var version = "dev"

func printBashCompletion() {
	script := `_gittag() {
    local cur prev commands flags
    COMPREPLY=()
    cur="${COMP_WORDS[COMP_CWORD]}"
    prev="${COMP_WORDS[COMP_CWORD-1]}"

    commands="init major minor patch pre-major pre-minor pre-patch pre release"
    flags="-v -version -n -dry-run -f -file -bash-completion"

    if [[ "${prev}" == "-file" ]]; then
        COMPREPLY=( $(compgen -f -- "${cur}") )
        return 0
    fi

    if [[ "${cur}" == -* ]]; then
        COMPREPLY=( $(compgen -W "${flags}" -- "${cur}") )
        return 0
    fi

    COMPREPLY=( $(compgen -W "${commands}" -- "${cur}") )
    return 0
}

complete -F _gittag gittag`
	fmt.Println(script)
}

func init() {
	flag.Usage = func() {
		progName := filepath.Base(os.Args[0])
		fmt.Fprintf(
			flag.CommandLine.Output(),
			`Usage: %s [flags] <command> [prerelease-id]

Commands:
  init                        Initialize with v0.0.0 tag (only if no version tag exists)
  major                       Bump major version (v1.2.3 -> v2.0.0)
  minor                       Bump minor version (v1.2.3 -> v1.3.0)
  patch                       Bump patch version (v1.2.3 -> v1.2.4)
  pre-major <prerelease-id>   Bump major and add prerelease (v1.2.3 -> v2.0.0-alpha)
  pre-minor <prerelease-id>   Bump minor and add prerelease (v1.2.3 -> v1.3.0-alpha)
  pre-patch <prerelease-id>   Bump patch and add prerelease (v1.2.3 -> v1.2.4-alpha)
  pre <prerelease-id>         Update prerelease identifier (v2.0.0-alpha -> v2.0.0-beta)
  release                     Remove prerelease to create release (v2.0.0-rc1 -> v2.0.0)

Version Source:
  By default, the current version is read from git tags. Use -f or -file to read
  from a file instead. When using file mode, the file is always updated, but git
  tags are only created for release versions and numbered prereleases (e.g.,
  alpha.1, rc1). Unnumbered prereleases (e.g., alpha, beta) do not create git
  tags.

Flags:
`,
			progName,
		)
		flag.PrintDefaults()
	}
}

var printVersion bool

func init() {
	const (
		defaultPrintVersion = false
		usage               = "Print version info"
	)
	flag.BoolVar(&printVersion, "version", defaultPrintVersion, usage)
	flag.BoolVar(&printVersion, "v", defaultPrintVersion, usage)
}

var dryRun bool

func init() {
	const (
		defaultDryRun = false
		usage         = "Dry run"
	)
	flag.BoolVar(&dryRun, "dry-run", defaultDryRun, usage)
	flag.BoolVar(&dryRun, "n", defaultDryRun, usage)
}

const defaultVersionFile = "VERSION"

var useFile bool
var versionFile string

func init() {
	flag.BoolVar(&useFile, "f", false, fmt.Sprintf("Read version from file '%s'", defaultVersionFile))
	flag.StringVar(&versionFile, "file", "", "Read version from specified file")
}

var bashCompletion bool

func init() {
	flag.BoolVar(&bashCompletion, "bash-completion", false, "Output bash completion script")
}

func main() {
	flag.Parse()

	if bashCompletion {
		printBashCompletion()
		os.Exit(0)
	}

	if printVersion {
		fmt.Println(version)
		os.Exit(0)
	}

	if flag.Arg(0) == "" {
		flag.Usage()
		os.Exit(1)
	}

	if versionFile != "" {
		useFile = true
	}
	if useFile && versionFile == "" {
		versionFile = defaultVersionFile
	}

	var opMode, prereleaseID string
	switch flag.Arg(0) {
	case Major, Minor, Patch, Release, Init:
		opMode = flag.Arg(0)
	case PreMajor, PreMinor, PrePatch, Pre:
		if flag.Arg(1) == "" {
			flag.Usage()
			os.Exit(1)
		}
		opMode = flag.Arg(0)
		prereleaseID = flag.Arg(1)
	default:
		fmt.Printf("Unknown command: %s\n", flag.Arg(0))
		flag.Usage()
		os.Exit(1)
	}

	if opMode == Init {
		if useFile {
			_, err := os.Stat(versionFile)
			if err == nil {
				fmt.Println("Version file already exists. Cannot init.")
				os.Exit(1)
			} else if !os.IsNotExist(err) {
				fmt.Println(err)
				os.Exit(1)
			}
		}
		hasTag, err := HasTag()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		if hasTag {
			fmt.Println("Found a version tag. Cannot init.")
			os.Exit(1)
		}
		initialVersion := Version{}
		if useFile {
			err := WriteVersionToFile(versionFile, initialVersion)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		}
		err = AddVersionTag(initialVersion)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		return
	}

	var currentTag string
	var err error
	if useFile {
		currentTag, err = GetVersionFromFile(versionFile)
		if err != nil {
			fmt.Println("Could not read version file.")
			fmt.Println(err)
			os.Exit(1)
		}
	} else {
		currentTag, err = GetCurrentTag()
		if err != nil {
			fmt.Println("Could not get current tag.")
			fmt.Println(err)
			os.Exit(1)
		}
	}

	currentVersion, err := ParseVersion(currentTag)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	var nextVersion Version
	switch opMode {
	case Major, Minor, Patch:
		if currentVersion.IsPrerelease() {
			fmt.Println(
				"Cannot bump version. Current version is pre-release." +
					" Do 'release' first.",
			)
			os.Exit(1)
		}
		nextVersion = currentVersion.Bump(opMode)
	case PreMajor, PreMinor, PrePatch:
		if currentVersion.IsPrerelease() {
			fmt.Println(
				"Cannot bump to pre-release version." +
					" Current version is already pre-release." +
					" Use 'pre' to update pre-release version.",
			)
			os.Exit(1)
		}
		nextVersion, err = currentVersion.Bump(opMode).SetPrerelease(prereleaseID)
		if err != nil {
			fmt.Println("Pre-release identifier is not valid per semver spec.")
			os.Exit(1)
		}
	case Pre:
		if !currentVersion.IsPrerelease() {
			fmt.Println("Cannot update pre-release identifier." +
				" Current version is not pre-release.",
			)
			os.Exit(1)
		}
		nextVersion, err = currentVersion.SetPrerelease(prereleaseID)
		if err != nil {
			fmt.Println("Pre-release identifier is not valid per semver spec.")
			os.Exit(1)
		}
	case Release:
		if !currentVersion.IsPrerelease() {
			fmt.Println(
				"Cannot bump to release version." +
					" Current version is not pre-release.",
			)
			os.Exit(1)
		}
		nextVersion = currentVersion.ClearPrerelease()
	}

	fmt.Printf("Will bump from %s to %s\n", currentVersion, nextVersion)
	if useFile && !nextVersion.ShouldCreateTag() {
		fmt.Println("No git tag will be created (prerelease without version number).")
	}

	if dryRun {
		fmt.Println("Dry run. Doing nothing.")
		os.Exit(0)
	}

	if useFile {
		err := WriteVersionToFile(versionFile, nextVersion)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		if !nextVersion.ShouldCreateTag() {
			return
		}
	}

	err = AddVersionTag(nextVersion)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
