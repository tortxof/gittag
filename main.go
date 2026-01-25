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
  from a file instead. When using file mode, the file is updated and committed
  before creating the git tag (ensuring the tag points to the correct version).
  Git tags are only created for release versions and numbered prereleases (e.g.,
  alpha.1, rc1). Unnumbered prereleases (e.g., alpha, beta) only update the
  file.

Flags:
`,
			progName,
		)
		flag.PrintDefaults()
	}
}

func main() {
	var printVersion bool
	const (
		defaultPrintVersion = false
		versionFlagUsage    = "Print version info"
	)
	flag.BoolVar(&printVersion, "version", defaultPrintVersion, versionFlagUsage)
	flag.BoolVar(&printVersion, "v", defaultPrintVersion, versionFlagUsage)

	var dryRun bool
	const (
		defaultDryRun   = false
		dryRunFlagUsage = "Dry run"
	)
	flag.BoolVar(&dryRun, "dry-run", defaultDryRun, dryRunFlagUsage)
	flag.BoolVar(&dryRun, "n", defaultDryRun, dryRunFlagUsage)

	const defaultVersionFile = "VERSION"
	var useFile bool
	var versionFile string
	flag.BoolVar(&useFile, "f", false, fmt.Sprintf("Read version from file '%s'", defaultVersionFile))
	flag.StringVar(&versionFile, "file", "", "Read version from specified file")

	var bashCompletion bool
	flag.BoolVar(&bashCompletion, "bash-completion", false, "Output bash completion script")

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

	options := Options{
		OpMode:       opMode,
		PrereleaseID: prereleaseID,
		UseFile:      useFile,
		VersionFile:  versionFile,
		DryRun:       dryRun,
	}

	operation, ok := operations[opMode]
	if !ok {
		fmt.Printf("Unknown operation mode: %s\n", opMode)
		flag.Usage()
		os.Exit(1)
	}

	err := operation(options)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
