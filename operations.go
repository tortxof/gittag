package main

import (
	"fmt"
	"os"
)

type Options struct {
	OpMode       string
	PrereleaseID string
	UseFile      bool
	VersionFile  string
	DryRun       bool
}

type Operation func(options Options) error

var operations = map[string]Operation{
	Major:    DoBump,
	Minor:    DoBump,
	Patch:    DoBump,
	PreMajor: DoBumpPre,
	PreMinor: DoBumpPre,
	PrePatch: DoBumpPre,
	Pre:      DoPre,
	Release:  DoRelease,
	Init:     DoInit,
}

func DoBump(o Options) error {
	currentVersion, err := GetCurrentVersion(o)
	if err != nil {
		return err
	}
	if currentVersion.IsPrerelease() {
		return fmt.Errorf(
			"Cannot bump version. Current version is pre-release." +
				" Do 'release' first.",
		)
	}
	nextVersion := currentVersion.Bump(o.OpMode)
	return ApplyVersion(currentVersion, nextVersion, o)
}

func DoBumpPre(o Options) error {
	currentVersion, err := GetCurrentVersion(o)
	if err != nil {
		return err
	}
	if currentVersion.IsPrerelease() {
		return fmt.Errorf(
			"Cannot bump to pre-release version." +
				" Current version is already pre-release." +
				" Use 'pre' to update pre-release version.",
		)
	}
	nextVersion, err := currentVersion.Bump(o.OpMode).SetPrerelease(o.PrereleaseID)
	if err != nil {
		return fmt.Errorf("Pre-release identifier is not valid per semver spec.")
	}
	return ApplyVersion(currentVersion, nextVersion, o)
}

func DoPre(o Options) error {
	currentVersion, err := GetCurrentVersion(o)
	if err != nil {
		return err
	}
	if !currentVersion.IsPrerelease() {
		return fmt.Errorf("Cannot update pre-release identifier." +
			" Current version is not pre-release.",
		)
	}
	nextVersion, err := currentVersion.SetPrerelease(o.PrereleaseID)
	if err != nil {
		return fmt.Errorf("Pre-release identifier is not valid per semver spec.")
	}
	return ApplyVersion(currentVersion, nextVersion, o)
}

func DoRelease(o Options) error {
	currentVersion, err := GetCurrentVersion(o)
	if err != nil {
		return err
	}
	if !currentVersion.IsPrerelease() {
		return fmt.Errorf(
			"Cannot bump to release version." +
				" Current version is not pre-release.",
		)
	}
	nextVersion := currentVersion.ClearPrerelease()
	return ApplyVersion(currentVersion, nextVersion, o)
}

func DoInit(o Options) error {
	if o.UseFile {
		_, err := os.Stat(o.VersionFile)
		if err == nil {
			return fmt.Errorf("Version file already exists. Cannot init.")
		} else if !os.IsNotExist(err) {
			return err
		}
	}
	hasTag, err := HasTag()
	if err != nil {
		return err
	}
	if hasTag {
		return fmt.Errorf("Found a version tag. Cannot init.")
	}
	initialVersion := Version{}
	if o.UseFile {
		err := WriteVersionToFile(o.VersionFile, initialVersion)
		if err != nil {
			return err
		}
	}
	err = AddVersionTag(initialVersion)
	if err != nil {
		return err
	}
	return nil
}

func GetCurrentVersion(o Options) (Version, error) {
	var currentTag string
	var err error
	if o.UseFile {
		currentTag, err = GetVersionFromFile(o.VersionFile)
		if err != nil {
			return Version{}, fmt.Errorf("Could not read version file.\n%s", err)
		}
	} else {
		currentTag, err = GetCurrentTag()
		if err != nil {
			return Version{}, fmt.Errorf("Could not get current tag.\n%s", err)
		}
	}

	currentVersion, err := ParseVersion(currentTag)
	if err != nil {
		return Version{}, err
	}
	return currentVersion, nil
}

func ApplyVersion(currentVersion Version, nextVersion Version, o Options) error {
	fmt.Printf("Will bump from %s to %s\n", currentVersion, nextVersion)
	if o.UseFile && !nextVersion.ShouldCreateTag() {
		fmt.Println("No git tag will be created (prerelease without version number).")
	}

	if o.DryRun {
		fmt.Println("Dry run. Doing nothing.")
		return nil
	}

	if o.UseFile {
		err := WriteVersionToFile(o.VersionFile, nextVersion)
		if err != nil {
			return err
		}
		if !nextVersion.ShouldCreateTag() {
			return nil
		}
	}

	err := AddVersionTag(nextVersion)
	if err != nil {
		return err
	}

	return nil
}
