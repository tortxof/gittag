package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

const (
	Major = "major"
	Minor = "minor"
	Patch = "patch"
)

type Version struct {
	Major int
	Minor int
	Patch int
}

func (v Version) Bump(level string) Version {
	switch level {
	case Major:
		return Version{
			Major: v.Major + 1,
			Minor: 0,
			Patch: 0,
		}
	case Minor:
		return Version{
			Major: v.Major,
			Minor: v.Minor + 1,
			Patch: 0,
		}
	case Patch:
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
	return fmt.Sprintf("v%d.%d.%d", v.Major, v.Minor, v.Patch)
}

func ParseVersion(tag string) (Version, error) {
	re := regexp.MustCompile(`^v(\d+)\.(\d+)\.(\d+)$`)
	matches := re.FindStringSubmatch(tag)

	if matches == nil {
		return Version{}, fmt.Errorf("tag does not match v{%s}.{%s}.{%s} format", Major, Minor, Patch)
	}

	var parts [3]int
	for i := range 3 {
		part, err := strconv.Atoi(matches[i+1])
		if err != nil {
			return Version{}, fmt.Errorf("tag part is not an integer: %q in %q", matches[i+1], tag)
		}
		parts[i] = part

	}

	return Version{Major: parts[0], Minor: parts[1], Patch: parts[2]}, nil
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

func AddVersionTag(v Version) error {
	cmd := exec.Command("git", "tag", v.String())
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git: %s", output)
	}
	return nil
}

func main() {
	if len(os.Args) != 2 {
		fmt.Printf("Usage: %s [%s|%s|%s]\n", filepath.Base(os.Args[0]), Major, Minor, Patch)
		os.Exit(1)
	}

	var opMode string
	switch os.Args[1] {
	case Major, Minor, Patch:
		opMode = os.Args[1]
	default:
		fmt.Printf("Argument must be '%s', '%s', or '%s'.\n", Major, Minor, Patch)
		os.Exit(1)
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

	nextVersion := currentVersion.Bump(opMode)

	fmt.Printf("Will bump from %s to %s", currentVersion.String(), nextVersion.String())

	err = AddVersionTag(nextVersion)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
