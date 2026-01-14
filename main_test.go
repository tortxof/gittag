package main

import (
	"fmt"
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
