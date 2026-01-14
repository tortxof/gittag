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

func (v Version) String() string {
	if v.Prerelease != "" {
		return fmt.Sprintf("v%d.%d.%d-%s", v.Major, v.Minor, v.Patch, v.Prerelease)
	}
	return fmt.Sprintf("v%d.%d.%d", v.Major, v.Minor, v.Patch)
}

func (v Version) IsPrerelease() bool {
	return v.Prerelease != ""
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

func init() {
	flag.Usage = func() {
		progName := filepath.Base(os.Args[0])
		fmt.Fprintf(
			flag.CommandLine.Output(),
			"Usage: %s [flags] <major|minor|patch>  Bump the version and create a new git tag\n",
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

func main() {
	flag.Parse()

	if printVersion {
		fmt.Println(version)
		os.Exit(0)
	}

	if flag.Arg(0) == "" {
		flag.Usage()
		os.Exit(1)
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
		hasTag, err := HasTag()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		if hasTag {
			fmt.Println("Found a version tag. Cannot init.")
			os.Exit(1)
		}
		err = AddVersionTag(Version{})
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		return
	}

	currentTag, err := GetCurrentTag()
	if err != nil {
		fmt.Println("Could not get current tag.")
		fmt.Println(err)
		os.Exit(1)
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
		nextVersion = currentVersion.Bump(opMode)
		nextVersion.Prerelease = prereleaseID
		nextVersion, err = ParseVersion(nextVersion.String())
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
		nextVersion = currentVersion
		nextVersion.Prerelease = prereleaseID
		nextVersion, err = ParseVersion(nextVersion.String())
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
		nextVersion = currentVersion
		nextVersion.Prerelease = ""
	}

	fmt.Printf("Will bump from %s to %s\n", currentVersion, nextVersion)

	if dryRun {
		fmt.Println("Dry run. Doing nothing.")
		os.Exit(0)
	}

	err = AddVersionTag(nextVersion)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
