package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

var version = "dev"

const (
	Init     = "init"
	Major    = "major"
	Minor    = "minor"
	Patch    = "patch"
	PreMajor = "pre-major"
	PreMinor = "pre-minor"
	PrePatch = "pre-patch"
	Pre      = "pre"
	Release  = "release"
)

type Version struct {
	Major      int
	Minor      int
	Patch      int
	Prerelease string
}

func (v Version) Bump(level string) Version {
	switch level {
	case Major, PreMajor:
		return Version{
			Major: v.Major + 1,
			Minor: 0,
			Patch: 0,
		}
	case Minor, PreMinor:
		return Version{
			Major: v.Major,
			Minor: v.Minor + 1,
			Patch: 0,
		}
	case Patch, PrePatch:
		return Version{
			Major: v.Major,
			Minor: v.Minor,
			Patch: v.Patch + 1,
		}
	default:
		panic(fmt.Sprintf("Invalid bump level: %s", level))
	}
}

func (v Version) SetPrerelease(prereleaseID string) (Version, error) {
	return ParseVersion(
		Version{
			Major:      v.Major,
			Minor:      v.Minor,
			Patch:      v.Patch,
			Prerelease: prereleaseID,
		}.String(),
	)
}

func (v Version) ClearPrerelease() Version {
	return Version{
		Major: v.Major,
		Minor: v.Minor,
		Patch: v.Patch,
	}
}

func (v Version) String() string {
	if v.IsPrerelease() {
		return fmt.Sprintf("v%d.%d.%d-%s", v.Major, v.Minor, v.Patch, v.Prerelease)
	}
	return fmt.Sprintf("v%d.%d.%d", v.Major, v.Minor, v.Patch)
}

func (v Version) IsPrerelease() bool {
	return v.Prerelease != ""
}

func (v Version) ShouldCreateTag() bool {
	if !v.IsPrerelease() {
		return true
	}
	lastChar := v.Prerelease[len(v.Prerelease)-1]
	return lastChar >= '0' && lastChar <= '9'
}

func ParseVersion(tag string) (Version, error) {
	re := regexp.MustCompile(`^v?(?P<major>0|[1-9]\d*)\.(?P<minor>0|[1-9]\d*)\.(?P<patch>0|[1-9]\d*)(?:-(?P<prerelease>(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\.(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\+(?P<buildmetadata>[0-9a-zA-Z-]+(?:\.[0-9a-zA-Z-]+)*))?$`)
	matches := re.FindStringSubmatch(tag)

	if matches == nil {
		return Version{}, fmt.Errorf("tag does not match semver format")
	}

	var parts [3]int
	for i := range 3 {
		part, err := strconv.Atoi(matches[i+1])
		if err != nil {
			return Version{}, fmt.Errorf("tag part is not an integer: %q in %q", matches[i+1], tag)
		}
		parts[i] = part
	}

	return Version{Major: parts[0], Minor: parts[1], Patch: parts[2], Prerelease: matches[4]}, nil
}

func GetVersionFromFile(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	if scanner.Scan() {
		return strings.TrimSpace(scanner.Text()), nil
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}

	return "", fmt.Errorf("file is empty")
}

func WriteVersionToFile(path string, v Version) error {
	content := v.String()[1:] + "\n"
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString(content)
	return err
}

func GetCurrentTag() (string, error) {
	cmd := exec.Command("git", "describe", "--tags", "--abbrev=0", "--match", "v*.*.*")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("git: %s", &stderr)
	}

	return strings.TrimSpace(stdout.String()), nil
}

func HasTag() (bool, error) {
	cmd := exec.Command("git", "tag", "--list")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		return false, fmt.Errorf("git: %s", &stderr)
	}

	scanner := bufio.NewScanner(&stdout)
	for scanner.Scan() {
		tag := strings.TrimSpace(scanner.Text())
		_, err := ParseVersion(tag)
		if err == nil {
			return true, nil
		}
	}
	if err := scanner.Err(); err != nil {
		return false, err
	}

	return false, nil
}

func AddVersionTag(v Version) error {
	cmd := exec.Command("git", "tag", v.String())
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git: %s", output)
	}
	return nil
}

func printBashCompletion() {
	script := `_gittag() {
    local cur prev commands flags
    COMPREPLY=()
    cur="${COMP_WORDS[COMP_CWORD]}"
    prev="${COMP_WORDS[COMP_CWORD-1]}"

    commands="init major minor patch pre-major pre-minor pre-patch pre release"
    flags="-v -version -n -dry-run -f -file -bash-completion"

    if [[ ${prev} == "-file" ]]; then
        COMPREPLY=( $(compgen -f -- ${cur}) )
        return 0
    fi

    if [[ ${cur} == -* ]]; then
        COMPREPLY=( $(compgen -W "${flags}" -- ${cur}) )
        return 0
    fi

    COMPREPLY=( $(compgen -W "${commands}" -- ${cur}) )
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
