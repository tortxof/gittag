package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestVersionBump(t *testing.T) {
	tests := []struct {
		name     string
		start    Version
		level    string
		want     Version
		panics   bool
		panicMsg string
	}{
		{
			name:  "patch",
			start: Version{Major: 1, Minor: 2, Patch: 3},
			level: Patch,
			want:  Version{Major: 1, Minor: 2, Patch: 4},
		},
		{
			name:  "minor",
			start: Version{Major: 2, Minor: 3, Patch: 4},
			level: Minor,
			want:  Version{Major: 2, Minor: 4, Patch: 0},
		},
		{
			name:  "major",
			start: Version{Major: 2, Minor: 3, Patch: 4},
			level: Major,
			want:  Version{Major: 3, Minor: 0, Patch: 0},
		},
		{
			name:  "pre-major",
			start: Version{Major: 1, Minor: 2, Patch: 3},
			level: PreMajor,
			want:  Version{Major: 2, Minor: 0, Patch: 0},
		},
		{
			name:  "pre-minor",
			start: Version{Major: 1, Minor: 2, Patch: 3},
			level: PreMinor,
			want:  Version{Major: 1, Minor: 3, Patch: 0},
		},
		{
			name:  "pre-patch",
			start: Version{Major: 1, Minor: 2, Patch: 3},
			level: PrePatch,
			want:  Version{Major: 1, Minor: 2, Patch: 4},
		},
		{
			name:     "invalid level panics",
			start:    Version{Major: 0, Minor: 0, Patch: 0},
			level:    "invalid",
			panics:   true,
			panicMsg: "Invalid bump level: invalid",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				recovered := recover()
				if tt.panics {
					if recovered == nil {
						t.Fatalf("expected panic for level %q", tt.level)
					}
					if tt.panicMsg != "" {
						if gotMsg := fmt.Sprint(recovered); gotMsg != tt.panicMsg {
							t.Fatalf("panic = %q, want %q", gotMsg, tt.panicMsg)
						}
					}
				} else if recovered != nil {
					t.Fatalf("did not expect panic for level %q: %v", tt.level, recovered)
				}
			}()

			got := tt.start.Bump(tt.level)
			if tt.panics {
				return
			}
			if got != tt.want {
				t.Fatalf("Bump(%s) = %+v, want %+v", tt.level, got, tt.want)
			}
		})
	}
}

func TestVersionString(t *testing.T) {
	tests := []struct {
		name string
		v    Version
		want string
	}{
		{
			name: "zero version",
			v:    Version{Major: 0, Minor: 0, Patch: 0},
			want: "v0.0.0",
		},
		{
			name: "mixed digits",
			v:    Version{Major: 10, Minor: 2, Patch: 45},
			want: "v10.2.45",
		},
		{
			name: "single digits",
			v:    Version{Major: 1, Minor: 2, Patch: 3},
			want: "v1.2.3",
		},
		{
			name: "with prerelease",
			v:    Version{Major: 1, Minor: 2, Patch: 3, Prerelease: "alpha.1"},
			want: "v1.2.3-alpha.1",
		},
	}

	for _, tt := range tests {
		if got := tt.v.String(); got != tt.want {
			t.Fatalf("String() = %q, want %q", got, tt.want)
		}
	}
}

func TestParseVersion(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    Version
		wantErr bool
	}{
		{
			name:  "with leading v",
			input: "v1.2.3",
			want:  Version{Major: 1, Minor: 2, Patch: 3},
		},
		{
			name:  "without leading v",
			input: "4.5.6",
			want:  Version{Major: 4, Minor: 5, Patch: 6},
		},
		{
			name:  "zero version",
			input: "0.0.0",
			want:  Version{Major: 0, Minor: 0, Patch: 0},
		},
		{
			name:  "with prerelease build",
			input: "7.8.9-beta.1+build.5",
			want:  Version{Major: 7, Minor: 8, Patch: 9, Prerelease: "beta.1"},
		},
		{
			name:  "complex prerelease and metadata",
			input: "0.0.1-rc.1.2+meta.data",
			want:  Version{Major: 0, Minor: 0, Patch: 1, Prerelease: "rc.1.2"},
		},
		{
			name:    "invalid format",
			input:   "not-a-tag",
			wantErr: true,
		},
		{
			name:    "missing patch",
			input:   "1.2",
			wantErr: true,
		},
		{
			name:    "leading zero major",
			input:   "01.2.3",
			wantErr: true,
		},
		{
			name:    "uppercase prefix",
			input:   "V1.2.3",
			wantErr: true,
		},
		{
			name:    "negative number",
			input:   "-1.2.3",
			wantErr: true,
		},
		{
			name:    "non numeric patch",
			input:   "1.2.x",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseVersion(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("ParseVersion(%q) expected error", tt.input)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseVersion(%q) unexpected error: %v", tt.input, err)
			}
			if got != tt.want {
				t.Fatalf("ParseVersion(%q) = %+v, want %+v", tt.input, got, tt.want)
			}
		})
	}
}

func TestGetVersionFromFile(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    string
		wantErr bool
	}{
		{
			name:    "simple version",
			content: "1.2.3\n",
			want:    "1.2.3",
		},
		{
			name:    "version with leading v",
			content: "v1.2.3\n",
			want:    "v1.2.3",
		},
		{
			name:    "version without trailing newline",
			content: "1.2.3",
			want:    "1.2.3",
		},
		{
			name:    "version with whitespace",
			content: "  1.2.3  \n",
			want:    "1.2.3",
		},
		{
			name:    "version with extra lines",
			content: "1.2.3\nextra\nlines\n",
			want:    "1.2.3",
		},
		{
			name:    "empty file",
			content: "",
			wantErr: true,
		},
		{
			name:    "only whitespace",
			content: "   \n",
			want:    "",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			path := filepath.Join(tmpDir, "VERSION")
			err := os.WriteFile(path, []byte(tt.content), 0644)
			if err != nil {
				t.Fatalf("failed to write test file: %v", err)
			}

			got, err := GetVersionFromFile(path)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("GetVersionFromFile() expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("GetVersionFromFile() unexpected error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("GetVersionFromFile() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestGetVersionFromFile_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "NONEXISTENT")
	_, err := GetVersionFromFile(path)
	if err == nil {
		t.Fatalf("GetVersionFromFile() expected error for nonexistent file")
	}
}

func TestWriteVersionToFile(t *testing.T) {
	tests := []struct {
		name    string
		version Version
		want    string
	}{
		{
			name:    "simple version",
			version: Version{Major: 1, Minor: 2, Patch: 3},
			want:    "1.2.3\n",
		},
		{
			name:    "zero version",
			version: Version{Major: 0, Minor: 0, Patch: 0},
			want:    "0.0.0\n",
		},
		{
			name:    "large numbers",
			version: Version{Major: 10, Minor: 20, Patch: 30},
			want:    "10.20.30\n",
		},
		{
			name:    "prerelease version",
			version: Version{Major: 1, Minor: 0, Patch: 0, Prerelease: "alpha.1"},
			want:    "1.0.0-alpha.1\n",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			path := filepath.Join(tmpDir, "VERSION")

			err := WriteVersionToFile(path, tt.version)
			if err != nil {
				t.Fatalf("WriteVersionToFile() unexpected error: %v", err)
			}

			content, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("failed to read test file: %v", err)
			}

			if string(content) != tt.want {
				t.Fatalf("WriteVersionToFile() wrote %q, want %q", string(content), tt.want)
			}
		})
	}
}

func TestWriteVersionToFile_Overwrite(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "VERSION")

	// Write initial content
	err := os.WriteFile(path, []byte("old content\n"), 0644)
	if err != nil {
		t.Fatalf("failed to write initial file: %v", err)
	}

	// Overwrite with version
	v := Version{Major: 2, Minor: 0, Patch: 0}
	err = WriteVersionToFile(path, v)
	if err != nil {
		t.Fatalf("WriteVersionToFile() unexpected error: %v", err)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read test file: %v", err)
	}

	want := "2.0.0\n"
	if string(content) != want {
		t.Fatalf("WriteVersionToFile() wrote %q, want %q", string(content), want)
	}
}

func TestGetRepoRoot(t *testing.T) {
	dir := t.TempDir()

	// Initialize git repo
	cmd := exec.Command("git", "init")
	cmd.Dir = dir
	if err := cmd.Run(); err != nil {
		t.Fatalf("git init failed: %v", err)
	}

	// Create a subdirectory
	subdir := filepath.Join(dir, "subdir", "nested")
	if err := os.MkdirAll(subdir, 0755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}

	// Save current working directory
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd failed: %v", err)
	}
	defer os.Chdir(origDir)

	// Change to subdirectory and call GetRepoRoot
	if err := os.Chdir(subdir); err != nil {
		t.Fatalf("chdir failed: %v", err)
	}

	root, err := GetRepoRoot()
	if err != nil {
		t.Fatalf("GetRepoRoot() unexpected error: %v", err)
	}

	// The returned root should match the original dir (normalized)
	if root != dir {
		t.Fatalf("GetRepoRoot() = %q, want %q", root, dir)
	}
}

func TestGetRepoRoot_NotGitRepo(t *testing.T) {
	dir := t.TempDir()

	// Save current working directory
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd failed: %v", err)
	}
	defer os.Chdir(origDir)

	// Change to non-git directory
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir failed: %v", err)
	}

	_, err = GetRepoRoot()
	if err == nil {
		t.Fatal("GetRepoRoot() expected error for non-git directory")
	}
}

func TestShouldCreateTag(t *testing.T) {
	tests := []struct {
		name    string
		version Version
		want    bool
	}{
		{
			name:    "release version",
			version: Version{Major: 1, Minor: 0, Patch: 0},
			want:    true,
		},
		{
			name:    "unnumbered prerelease alpha",
			version: Version{Major: 1, Minor: 0, Patch: 0, Prerelease: "alpha"},
			want:    false,
		},
		{
			name:    "unnumbered prerelease beta",
			version: Version{Major: 1, Minor: 0, Patch: 0, Prerelease: "beta"},
			want:    false,
		},
		{
			name:    "unnumbered prerelease rc",
			version: Version{Major: 1, Minor: 0, Patch: 0, Prerelease: "rc"},
			want:    false,
		},
		{
			name:    "numbered prerelease alpha.1",
			version: Version{Major: 1, Minor: 0, Patch: 0, Prerelease: "alpha.1"},
			want:    true,
		},
		{
			name:    "numbered prerelease beta.2",
			version: Version{Major: 1, Minor: 0, Patch: 0, Prerelease: "beta.2"},
			want:    true,
		},
		{
			name:    "numbered prerelease rc1",
			version: Version{Major: 1, Minor: 0, Patch: 0, Prerelease: "rc1"},
			want:    true,
		},
		{
			name:    "numbered prerelease rc10",
			version: Version{Major: 1, Minor: 0, Patch: 0, Prerelease: "rc10"},
			want:    true,
		},
		{
			name:    "zero version",
			version: Version{Major: 0, Minor: 0, Patch: 0},
			want:    true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			got := tt.version.ShouldCreateTag()
			if got != tt.want {
				t.Fatalf("ShouldCreateTag() = %v, want %v", got, tt.want)
			}
		})
	}
}
