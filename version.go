package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
)

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
