package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

var binaryPath string

func TestMain(m *testing.M) {
	// Build the binary once for all integration tests
	tmpDir, err := os.MkdirTemp("", "gittag-test-bin")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(tmpDir)

	binaryPath = filepath.Join(tmpDir, "gittag")
	cmd := exec.Command("go", "build", "-o", binaryPath, ".")
	if output, err := cmd.CombinedOutput(); err != nil {
		panic(string(output))
	}

	os.Exit(m.Run())
}

// setupGitRepo creates a temp directory with an initialized git repo
// and optionally adds a starting tag
func setupGitRepo(t *testing.T, startingTag string) string {
	t.Helper()

	dir := t.TempDir()

	// Initialize git repo
	runGit(t, dir, "init")
	runGit(t, dir, "config", "user.email", "test@test.com")
	runGit(t, dir, "config", "user.name", "Test User")

	// Create an initial commit (required for tags)
	testFile := filepath.Join(dir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}
	runGit(t, dir, "add", ".")
	runGit(t, dir, "commit", "-m", "initial commit")

	// Add starting tag if specified
	if startingTag != "" {
		runGit(t, dir, "tag", startingTag)

		// Create another commit so the new tag will be on a different commit
		// This is needed because git describe returns the first tag when
		// multiple tags point to the same commit
		if err := os.WriteFile(testFile, []byte("test2"), 0644); err != nil {
			t.Fatal(err)
		}
		runGit(t, dir, "add", ".")
		runGit(t, dir, "commit", "-m", "second commit")
	}

	return dir
}

func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v failed: %v\n%s", args, err, output)
	}
}

func runGittag(t *testing.T, dir string, args ...string) (string, error) {
	t.Helper()
	cmd := exec.Command(binaryPath, args...)
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	return string(output), err
}

func getLatestTag(t *testing.T, dir string) string {
	t.Helper()
	cmd := exec.Command("git", "describe", "--tags", "--abbrev=0", "--match", "v*.*.*")
	cmd.Dir = dir
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("failed to get latest tag: %v", err)
	}
	return strings.TrimSpace(string(output))
}

func TestIntegrationInit(t *testing.T) {
	dir := setupGitRepo(t, "") // No starting tag for init

	_, err := runGittag(t, dir, "init")
	if err != nil {
		t.Fatalf("gittag init failed: %v", err)
	}

	tag := getLatestTag(t, dir)
	if tag != "v0.0.0" {
		t.Fatalf("expected v0.0.0, got %s", tag)
	}
}

func TestIntegrationMajor(t *testing.T) {
	dir := setupGitRepo(t, "v1.2.3")

	_, err := runGittag(t, dir, "major")
	if err != nil {
		t.Fatalf("gittag major failed: %v", err)
	}

	tag := getLatestTag(t, dir)
	if tag != "v2.0.0" {
		t.Fatalf("expected v2.0.0, got %s", tag)
	}
}

func TestIntegrationMinor(t *testing.T) {
	dir := setupGitRepo(t, "v1.2.3")

	_, err := runGittag(t, dir, "minor")
	if err != nil {
		t.Fatalf("gittag minor failed: %v", err)
	}

	tag := getLatestTag(t, dir)
	if tag != "v1.3.0" {
		t.Fatalf("expected v1.3.0, got %s", tag)
	}
}

func TestIntegrationPatch(t *testing.T) {
	dir := setupGitRepo(t, "v1.2.3")

	_, err := runGittag(t, dir, "patch")
	if err != nil {
		t.Fatalf("gittag patch failed: %v", err)
	}

	tag := getLatestTag(t, dir)
	if tag != "v1.2.4" {
		t.Fatalf("expected v1.2.4, got %s", tag)
	}
}

func TestIntegrationPreMajor(t *testing.T) {
	dir := setupGitRepo(t, "v1.2.3")

	_, err := runGittag(t, dir, "pre-major", "alpha.1")
	if err != nil {
		t.Fatalf("gittag pre-major failed: %v", err)
	}

	tag := getLatestTag(t, dir)
	if tag != "v2.0.0-alpha.1" {
		t.Fatalf("expected v2.0.0-alpha.1, got %s", tag)
	}
}

func TestIntegrationPreMinor(t *testing.T) {
	dir := setupGitRepo(t, "v1.2.3")

	_, err := runGittag(t, dir, "pre-minor", "beta.1")
	if err != nil {
		t.Fatalf("gittag pre-minor failed: %v", err)
	}

	tag := getLatestTag(t, dir)
	if tag != "v1.3.0-beta.1" {
		t.Fatalf("expected v1.3.0-beta.1, got %s", tag)
	}
}

func TestIntegrationPrePatch(t *testing.T) {
	dir := setupGitRepo(t, "v1.2.3")

	_, err := runGittag(t, dir, "pre-patch", "rc.1")
	if err != nil {
		t.Fatalf("gittag pre-patch failed: %v", err)
	}

	tag := getLatestTag(t, dir)
	if tag != "v1.2.4-rc.1" {
		t.Fatalf("expected v1.2.4-rc.1, got %s", tag)
	}
}

func TestIntegrationPre(t *testing.T) {
	dir := setupGitRepo(t, "v1.3.0-alpha.1")

	_, err := runGittag(t, dir, "pre", "alpha.2")
	if err != nil {
		t.Fatalf("gittag pre failed: %v", err)
	}

	tag := getLatestTag(t, dir)
	if tag != "v1.3.0-alpha.2" {
		t.Fatalf("expected v1.3.0-alpha.2, got %s", tag)
	}
}

func TestIntegrationRelease(t *testing.T) {
	dir := setupGitRepo(t, "v1.3.0-beta.2")

	_, err := runGittag(t, dir, "release")
	if err != nil {
		t.Fatalf("gittag release failed: %v", err)
	}

	tag := getLatestTag(t, dir)
	if tag != "v1.3.0" {
		t.Fatalf("expected v1.3.0, got %s", tag)
	}
}

// Error case tests

func TestIntegrationInitWithExistingTag(t *testing.T) {
	dir := setupGitRepo(t, "v1.0.0")

	output, err := runGittag(t, dir, "init")
	if err == nil {
		t.Fatal("expected gittag init to fail when tag exists")
	}
	if !strings.Contains(output, "Cannot init") {
		t.Fatalf("expected error message about cannot init, got: %s", output)
	}
}

func TestIntegrationMajorOnPrerelease(t *testing.T) {
	dir := setupGitRepo(t, "v1.2.3-alpha.1")

	output, err := runGittag(t, dir, "major")
	if err == nil {
		t.Fatal("expected gittag major to fail on pre-release version")
	}
	if !strings.Contains(output, "pre-release") {
		t.Fatalf("expected error message about pre-release, got: %s", output)
	}
}

func TestIntegrationMinorOnPrerelease(t *testing.T) {
	dir := setupGitRepo(t, "v1.2.3-beta.1")

	output, err := runGittag(t, dir, "minor")
	if err == nil {
		t.Fatal("expected gittag minor to fail on pre-release version")
	}
	if !strings.Contains(output, "pre-release") {
		t.Fatalf("expected error message about pre-release, got: %s", output)
	}
}

func TestIntegrationPatchOnPrerelease(t *testing.T) {
	dir := setupGitRepo(t, "v1.2.3-rc.1")

	output, err := runGittag(t, dir, "patch")
	if err == nil {
		t.Fatal("expected gittag patch to fail on pre-release version")
	}
	if !strings.Contains(output, "pre-release") {
		t.Fatalf("expected error message about pre-release, got: %s", output)
	}
}

func TestIntegrationPreMajorOnPrerelease(t *testing.T) {
	dir := setupGitRepo(t, "v1.2.3-alpha.1")

	output, err := runGittag(t, dir, "pre-major", "beta.1")
	if err == nil {
		t.Fatal("expected gittag pre-major to fail on pre-release version")
	}
	if !strings.Contains(output, "pre-release") {
		t.Fatalf("expected error message about pre-release, got: %s", output)
	}
}

func TestIntegrationPreMinorOnPrerelease(t *testing.T) {
	dir := setupGitRepo(t, "v1.2.3-alpha.1")

	output, err := runGittag(t, dir, "pre-minor", "beta.1")
	if err == nil {
		t.Fatal("expected gittag pre-minor to fail on pre-release version")
	}
	if !strings.Contains(output, "pre-release") {
		t.Fatalf("expected error message about pre-release, got: %s", output)
	}
}

func TestIntegrationPrePatchOnPrerelease(t *testing.T) {
	dir := setupGitRepo(t, "v1.2.3-alpha.1")

	output, err := runGittag(t, dir, "pre-patch", "beta.1")
	if err == nil {
		t.Fatal("expected gittag pre-patch to fail on pre-release version")
	}
	if !strings.Contains(output, "pre-release") {
		t.Fatalf("expected error message about pre-release, got: %s", output)
	}
}

func TestIntegrationPreOnNonPrerelease(t *testing.T) {
	dir := setupGitRepo(t, "v1.2.3")

	output, err := runGittag(t, dir, "pre", "alpha.1")
	if err == nil {
		t.Fatal("expected gittag pre to fail on non-pre-release version")
	}
	if !strings.Contains(output, "not pre-release") {
		t.Fatalf("expected error message about not pre-release, got: %s", output)
	}
}

func TestIntegrationReleaseOnNonPrerelease(t *testing.T) {
	dir := setupGitRepo(t, "v1.2.3")

	output, err := runGittag(t, dir, "release")
	if err == nil {
		t.Fatal("expected gittag release to fail on non-pre-release version")
	}
	if !strings.Contains(output, "not pre-release") {
		t.Fatalf("expected error message about not pre-release, got: %s", output)
	}
}

func TestIntegrationInvalidPrereleaseIdentifier(t *testing.T) {
	dir := setupGitRepo(t, "v1.2.3")

	output, err := runGittag(t, dir, "pre-minor", "01.invalid")
	if err == nil {
		t.Fatal("expected gittag pre-minor to fail with invalid prerelease identifier")
	}
	if !strings.Contains(output, "not valid") {
		t.Fatalf("expected error message about invalid identifier, got: %s", output)
	}
}
