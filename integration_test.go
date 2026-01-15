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

// File-based version tests

// setupGitRepoWithVersionFile creates a temp directory with an initialized git repo
// and a VERSION file with the specified version (without "v" prefix)
func setupGitRepoWithVersionFile(t *testing.T, version string) string {
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

	// Create VERSION file if version specified
	if version != "" {
		versionFile := filepath.Join(dir, "VERSION")
		if err := os.WriteFile(versionFile, []byte(version+"\n"), 0644); err != nil {
			t.Fatal(err)
		}

		// Add a git tag matching the version
		runGit(t, dir, "tag", "v"+version)

		// Create another commit so new tags will be on a different commit
		if err := os.WriteFile(testFile, []byte("test2"), 0644); err != nil {
			t.Fatal(err)
		}
		runGit(t, dir, "add", ".")
		runGit(t, dir, "commit", "-m", "second commit")
	}

	return dir
}

func getVersionFileContent(t *testing.T, dir string) string {
	t.Helper()
	content, err := os.ReadFile(filepath.Join(dir, "VERSION"))
	if err != nil {
		t.Fatalf("failed to read VERSION file: %v", err)
	}
	return strings.TrimSpace(string(content))
}

func getCustomVersionFileContent(t *testing.T, dir, filename string) string {
	t.Helper()
	content, err := os.ReadFile(filepath.Join(dir, filename))
	if err != nil {
		t.Fatalf("failed to read %s file: %v", filename, err)
	}
	return strings.TrimSpace(string(content))
}

func TestIntegrationFileInit(t *testing.T) {
	dir := setupGitRepoWithVersionFile(t, "") // No VERSION file

	_, err := runGittag(t, dir, "-f", "init")
	if err != nil {
		t.Fatalf("gittag -f init failed: %v", err)
	}

	// Check VERSION file was created
	version := getVersionFileContent(t, dir)
	if version != "0.0.0" {
		t.Fatalf("expected VERSION file to contain 0.0.0, got %s", version)
	}

	// Check git tag was also created
	tag := getLatestTag(t, dir)
	if tag != "v0.0.0" {
		t.Fatalf("expected v0.0.0 tag, got %s", tag)
	}
}

func TestIntegrationFileInitCustomFile(t *testing.T) {
	dir := setupGitRepoWithVersionFile(t, "") // No VERSION file

	_, err := runGittag(t, dir, "-file", "MYVERSION", "init")
	if err != nil {
		t.Fatalf("gittag -file MYVERSION init failed: %v", err)
	}

	// Check custom file was created
	version := getCustomVersionFileContent(t, dir, "MYVERSION")
	if version != "0.0.0" {
		t.Fatalf("expected MYVERSION file to contain 0.0.0, got %s", version)
	}

	// Check git tag was also created
	tag := getLatestTag(t, dir)
	if tag != "v0.0.0" {
		t.Fatalf("expected v0.0.0 tag, got %s", tag)
	}
}

func TestIntegrationFileInitWithExistingFile(t *testing.T) {
	dir := setupGitRepoWithVersionFile(t, "1.0.0") // VERSION file exists

	output, err := runGittag(t, dir, "-f", "init")
	if err == nil {
		t.Fatal("expected gittag -f init to fail when VERSION file exists")
	}
	if !strings.Contains(output, "already exists") {
		t.Fatalf("expected error message about file already exists, got: %s", output)
	}
}

func TestIntegrationFilePatch(t *testing.T) {
	dir := setupGitRepoWithVersionFile(t, "1.2.3")

	_, err := runGittag(t, dir, "-f", "patch")
	if err != nil {
		t.Fatalf("gittag -f patch failed: %v", err)
	}

	// Check VERSION file was updated
	version := getVersionFileContent(t, dir)
	if version != "1.2.4" {
		t.Fatalf("expected VERSION file to contain 1.2.4, got %s", version)
	}

	// Check git tag was also created
	tag := getLatestTag(t, dir)
	if tag != "v1.2.4" {
		t.Fatalf("expected v1.2.4 tag, got %s", tag)
	}
}

func TestIntegrationFileMinor(t *testing.T) {
	dir := setupGitRepoWithVersionFile(t, "1.2.3")

	_, err := runGittag(t, dir, "-f", "minor")
	if err != nil {
		t.Fatalf("gittag -f minor failed: %v", err)
	}

	// Check VERSION file was updated
	version := getVersionFileContent(t, dir)
	if version != "1.3.0" {
		t.Fatalf("expected VERSION file to contain 1.3.0, got %s", version)
	}

	// Check git tag was also created
	tag := getLatestTag(t, dir)
	if tag != "v1.3.0" {
		t.Fatalf("expected v1.3.0 tag, got %s", tag)
	}
}

func TestIntegrationFileMajor(t *testing.T) {
	dir := setupGitRepoWithVersionFile(t, "1.2.3")

	_, err := runGittag(t, dir, "-f", "major")
	if err != nil {
		t.Fatalf("gittag -f major failed: %v", err)
	}

	// Check VERSION file was updated
	version := getVersionFileContent(t, dir)
	if version != "2.0.0" {
		t.Fatalf("expected VERSION file to contain 2.0.0, got %s", version)
	}

	// Check git tag was also created
	tag := getLatestTag(t, dir)
	if tag != "v2.0.0" {
		t.Fatalf("expected v2.0.0 tag, got %s", tag)
	}
}

func TestIntegrationFileCustomPath(t *testing.T) {
	dir := setupGitRepoWithVersionFile(t, "")

	// Create a custom version file
	customFile := filepath.Join(dir, "my-version.txt")
	if err := os.WriteFile(customFile, []byte("2.0.0\n"), 0644); err != nil {
		t.Fatal(err)
	}

	// Also need a matching git tag
	runGit(t, dir, "tag", "v2.0.0")

	// Create another commit
	testFile := filepath.Join(dir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test3"), 0644); err != nil {
		t.Fatal(err)
	}
	runGit(t, dir, "add", ".")
	runGit(t, dir, "commit", "-m", "third commit")

	_, err := runGittag(t, dir, "-file", "my-version.txt", "patch")
	if err != nil {
		t.Fatalf("gittag -file my-version.txt patch failed: %v", err)
	}

	// Check custom file was updated
	version := getCustomVersionFileContent(t, dir, "my-version.txt")
	if version != "2.0.1" {
		t.Fatalf("expected my-version.txt to contain 2.0.1, got %s", version)
	}

	// Check git tag was created
	tag := getLatestTag(t, dir)
	if tag != "v2.0.1" {
		t.Fatalf("expected v2.0.1 tag, got %s", tag)
	}
}

func TestIntegrationFileDryRun(t *testing.T) {
	dir := setupGitRepoWithVersionFile(t, "1.2.3")

	output, err := runGittag(t, dir, "-f", "-n", "patch")
	if err != nil {
		t.Fatalf("gittag -f -n patch failed: %v", err)
	}

	// Check output shows what would happen
	if !strings.Contains(output, "v1.2.3") || !strings.Contains(output, "v1.2.4") {
		t.Fatalf("expected dry-run output to show version change, got: %s", output)
	}

	// Check VERSION file was NOT updated
	version := getVersionFileContent(t, dir)
	if version != "1.2.3" {
		t.Fatalf("expected VERSION file to remain 1.2.3 after dry-run, got %s", version)
	}

	// Check no new tag was created
	tag := getLatestTag(t, dir)
	if tag != "v1.2.3" {
		t.Fatalf("expected tag to remain v1.2.3 after dry-run, got %s", tag)
	}
}

func TestIntegrationFilePreMinor(t *testing.T) {
	dir := setupGitRepoWithVersionFile(t, "1.2.3")

	_, err := runGittag(t, dir, "-f", "pre-minor", "alpha.1")
	if err != nil {
		t.Fatalf("gittag -f pre-minor failed: %v", err)
	}

	// Check VERSION file was updated with prerelease
	version := getVersionFileContent(t, dir)
	if version != "1.3.0-alpha.1" {
		t.Fatalf("expected VERSION file to contain 1.3.0-alpha.1, got %s", version)
	}

	// Check git tag was created
	tag := getLatestTag(t, dir)
	if tag != "v1.3.0-alpha.1" {
		t.Fatalf("expected v1.3.0-alpha.1 tag, got %s", tag)
	}
}

func TestIntegrationFileRelease(t *testing.T) {
	dir := setupGitRepoWithVersionFile(t, "1.3.0-beta.2")

	// Need a matching git tag for the prerelease
	runGit(t, dir, "tag", "-d", "v1.3.0-beta.2") // Remove auto-created tag first
	runGit(t, dir, "tag", "v1.3.0-beta.2")

	// Create another commit
	testFile := filepath.Join(dir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test3"), 0644); err != nil {
		t.Fatal(err)
	}
	runGit(t, dir, "add", ".")
	runGit(t, dir, "commit", "-m", "third commit")

	_, err := runGittag(t, dir, "-f", "release")
	if err != nil {
		t.Fatalf("gittag -f release failed: %v", err)
	}

	// Check VERSION file was updated
	version := getVersionFileContent(t, dir)
	if version != "1.3.0" {
		t.Fatalf("expected VERSION file to contain 1.3.0, got %s", version)
	}

	// Check git tag was created
	tag := getLatestTag(t, dir)
	if tag != "v1.3.0" {
		t.Fatalf("expected v1.3.0 tag, got %s", tag)
	}
}

func TestIntegrationFileMissingFile(t *testing.T) {
	dir := setupGitRepoWithVersionFile(t, "") // No VERSION file

	output, err := runGittag(t, dir, "-f", "patch")
	if err == nil {
		t.Fatal("expected gittag -f patch to fail when VERSION file doesn't exist")
	}
	if !strings.Contains(output, "Could not read version file") {
		t.Fatalf("expected error about reading version file, got: %s", output)
	}
}

// Test that unnumbered prerelease does NOT create a git tag
func TestIntegrationFilePreMinorNoTag(t *testing.T) {
	dir := setupGitRepoWithVersionFile(t, "1.2.3")

	output, err := runGittag(t, dir, "-f", "pre-minor", "alpha")
	if err != nil {
		t.Fatalf("gittag -f pre-minor failed: %v", err)
	}

	// Check output mentions no tag will be created
	if !strings.Contains(output, "No git tag will be created") {
		t.Fatalf("expected output to mention no git tag, got: %s", output)
	}

	// Check VERSION file was updated
	version := getVersionFileContent(t, dir)
	if version != "1.3.0-alpha" {
		t.Fatalf("expected VERSION file to contain 1.3.0-alpha, got %s", version)
	}

	// Check git tag was NOT created (should still be v1.2.3)
	tag := getLatestTag(t, dir)
	if tag != "v1.2.3" {
		t.Fatalf("expected tag to remain v1.2.3 (no new tag), got %s", tag)
	}
}

// Test that numbered prerelease DOES create a git tag
func TestIntegrationFilePreNumberedTag(t *testing.T) {
	dir := setupGitRepoWithVersionFile(t, "1.3.0-alpha")

	// Need a matching git tag - but since alpha doesn't create a tag,
	// we need to set up the scenario correctly
	// The VERSION file has 1.3.0-alpha, but there's no v1.3.0-alpha tag
	// So the latest tag is still v1.3.0-alpha from setupGitRepoWithVersionFile
	// Actually, setupGitRepoWithVersionFile creates a tag, so we need to remove it

	// Remove the auto-created tag and don't create one (simulating unnumbered prerelease)
	runGit(t, dir, "tag", "-d", "v1.3.0-alpha")

	// Create another commit
	testFile := filepath.Join(dir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test3"), 0644); err != nil {
		t.Fatal(err)
	}
	runGit(t, dir, "add", ".")
	runGit(t, dir, "commit", "-m", "third commit")

	// Now bump to alpha.1 (numbered)
	_, err := runGittag(t, dir, "-f", "pre", "alpha.1")
	if err != nil {
		t.Fatalf("gittag -f pre failed: %v", err)
	}

	// Check VERSION file was updated
	version := getVersionFileContent(t, dir)
	if version != "1.3.0-alpha.1" {
		t.Fatalf("expected VERSION file to contain 1.3.0-alpha.1, got %s", version)
	}

	// Check git tag WAS created
	tag := getLatestTag(t, dir)
	if tag != "v1.3.0-alpha.1" {
		t.Fatalf("expected v1.3.0-alpha.1 tag, got %s", tag)
	}
}

// Test that pre with unnumbered identifier does NOT create a tag
func TestIntegrationFilePreUnnumberedNoTag(t *testing.T) {
	dir := setupGitRepoWithVersionFile(t, "1.3.0-alpha.1")

	// Create another commit
	testFile := filepath.Join(dir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test3"), 0644); err != nil {
		t.Fatal(err)
	}
	runGit(t, dir, "add", ".")
	runGit(t, dir, "commit", "-m", "third commit")

	output, err := runGittag(t, dir, "-f", "pre", "beta")
	if err != nil {
		t.Fatalf("gittag -f pre failed: %v", err)
	}

	// Check output mentions no tag will be created
	if !strings.Contains(output, "No git tag will be created") {
		t.Fatalf("expected output to mention no git tag, got: %s", output)
	}

	// Check VERSION file was updated
	version := getVersionFileContent(t, dir)
	if version != "1.3.0-beta" {
		t.Fatalf("expected VERSION file to contain 1.3.0-beta, got %s", version)
	}

	// Check git tag was NOT created (should still be v1.3.0-alpha.1)
	tag := getLatestTag(t, dir)
	if tag != "v1.3.0-alpha.1" {
		t.Fatalf("expected tag to remain v1.3.0-alpha.1 (no new tag), got %s", tag)
	}
}

// Test the full workflow from the issue description
func TestIntegrationFileWorkflow(t *testing.T) {
	dir := setupGitRepoWithVersionFile(t, "3.4.0")

	makeCommit := func() {
		testFile := filepath.Join(dir, "test.txt")
		content, _ := os.ReadFile(testFile)
		os.WriteFile(testFile, append(content, []byte("x")...), 0644)
		runGit(t, dir, "add", ".")
		runGit(t, dir, "commit", "-m", "commit")
	}

	// Step 1: pre-minor alpha (no tag)
	makeCommit()
	_, err := runGittag(t, dir, "-f", "pre-minor", "alpha")
	if err != nil {
		t.Fatalf("pre-minor alpha failed: %v", err)
	}
	if v := getVersionFileContent(t, dir); v != "3.5.0-alpha" {
		t.Fatalf("expected 3.5.0-alpha, got %s", v)
	}
	if tag := getLatestTag(t, dir); tag != "v3.4.0" {
		t.Fatalf("expected tag to remain v3.4.0, got %s", tag)
	}

	// Step 2: pre alpha.1 (tag created)
	makeCommit()
	_, err = runGittag(t, dir, "-f", "pre", "alpha.1")
	if err != nil {
		t.Fatalf("pre alpha.1 failed: %v", err)
	}
	if v := getVersionFileContent(t, dir); v != "3.5.0-alpha.1" {
		t.Fatalf("expected 3.5.0-alpha.1, got %s", v)
	}
	if tag := getLatestTag(t, dir); tag != "v3.5.0-alpha.1" {
		t.Fatalf("expected v3.5.0-alpha.1 tag, got %s", tag)
	}

	// Step 3: pre beta (no tag)
	makeCommit()
	_, err = runGittag(t, dir, "-f", "pre", "beta")
	if err != nil {
		t.Fatalf("pre beta failed: %v", err)
	}
	if v := getVersionFileContent(t, dir); v != "3.5.0-beta" {
		t.Fatalf("expected 3.5.0-beta, got %s", v)
	}
	if tag := getLatestTag(t, dir); tag != "v3.5.0-alpha.1" {
		t.Fatalf("expected tag to remain v3.5.0-alpha.1, got %s", tag)
	}

	// Step 4: pre beta.1 (tag created)
	makeCommit()
	_, err = runGittag(t, dir, "-f", "pre", "beta.1")
	if err != nil {
		t.Fatalf("pre beta.1 failed: %v", err)
	}
	if v := getVersionFileContent(t, dir); v != "3.5.0-beta.1" {
		t.Fatalf("expected 3.5.0-beta.1, got %s", v)
	}
	if tag := getLatestTag(t, dir); tag != "v3.5.0-beta.1" {
		t.Fatalf("expected v3.5.0-beta.1 tag, got %s", tag)
	}

	// Step 5: pre rc1 (tag created - ends with digit)
	makeCommit()
	_, err = runGittag(t, dir, "-f", "pre", "rc1")
	if err != nil {
		t.Fatalf("pre rc1 failed: %v", err)
	}
	if v := getVersionFileContent(t, dir); v != "3.5.0-rc1" {
		t.Fatalf("expected 3.5.0-rc1, got %s", v)
	}
	if tag := getLatestTag(t, dir); tag != "v3.5.0-rc1" {
		t.Fatalf("expected v3.5.0-rc1 tag, got %s", tag)
	}

	// Step 6: release (tag created)
	makeCommit()
	_, err = runGittag(t, dir, "-f", "release")
	if err != nil {
		t.Fatalf("release failed: %v", err)
	}
	if v := getVersionFileContent(t, dir); v != "3.5.0" {
		t.Fatalf("expected 3.5.0, got %s", v)
	}
	if tag := getLatestTag(t, dir); tag != "v3.5.0" {
		t.Fatalf("expected v3.5.0 tag, got %s", tag)
	}
}
