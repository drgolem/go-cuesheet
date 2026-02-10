package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestNormalizeSingleFile tests normalizing a single CUE file
func TestNormalizeSingleFile(t *testing.T) {
	// Create a temporary directory
	tmpDir := t.TempDir()

	// Create a test CUE file with wrong extension
	cueContent := `REM GENRE "Rock"
PERFORMER "Test Artist"
TITLE "Test Album"
FILE "test.wav" WAVE
  TRACK 01 AUDIO
    TITLE "Test Track"
    INDEX 01 00:00:00
`
	cuePath := filepath.Join(tmpDir, "test.cue")
	if err := os.WriteFile(cuePath, []byte(cueContent), 0644); err != nil {
		t.Fatalf("Failed to create test CUE file: %v", err)
	}

	// Create a test audio file with different extension
	audioPath := filepath.Join(tmpDir, "test.flac")
	if err := os.WriteFile(audioPath, []byte("dummy audio"), 0644); err != nil {
		t.Fatalf("Failed to create test audio file: %v", err)
	}

	// Process the CUE file (will backup and replace)
	changes := processCueFile(cuePath, "", false, false, false)

	if changes == 0 {
		t.Error("Expected changes but got 0")
	}

	// Check that backup was created
	backupPath := cuePath + ".bak"
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		t.Error("Backup file was not created")
	}

	// Read normalized content from original file location
	content, err := os.ReadFile(cuePath)
	if err != nil {
		t.Fatalf("Failed to read normalized file: %v", err)
	}

	// Check that the extension was corrected
	if !strings.Contains(string(content), "test.flac") {
		t.Error("Expected 'test.flac' in normalized content")
	}
	if strings.Contains(string(content), "test.wav") {
		t.Error("Should not contain 'test.wav' in normalized content")
	}

	// Verify backup contains original content
	backupContent, err := os.ReadFile(backupPath)
	if err != nil {
		t.Fatalf("Failed to read backup file: %v", err)
	}
	if !strings.Contains(string(backupContent), "test.wav") {
		t.Error("Expected backup to contain original 'test.wav'")
	}
}

// TestDryRunMode tests that dry-run mode doesn't create files
func TestDryRunMode(t *testing.T) {
	tmpDir := t.TempDir()

	cueContent := `FILE "test.wav" WAVE
  TRACK 01 AUDIO
    INDEX 01 00:00:00
`
	cuePath := filepath.Join(tmpDir, "test.cue")
	if err := os.WriteFile(cuePath, []byte(cueContent), 0644); err != nil {
		t.Fatalf("Failed to create test CUE file: %v", err)
	}

	audioPath := filepath.Join(tmpDir, "test.flac")
	if err := os.WriteFile(audioPath, []byte("dummy"), 0644); err != nil {
		t.Fatalf("Failed to create test audio file: %v", err)
	}

	// Process in dry-run mode
	changes := processCueFile(cuePath, "", true, false, false)

	if changes == 0 {
		t.Error("Expected changes detection in dry-run mode")
	}

	// Check that backup was NOT created
	backupPath := cuePath + ".bak"
	if _, err := os.Stat(backupPath); !os.IsNotExist(err) {
		t.Error("Backup file should not be created in dry-run mode")
	}

	// Check that original file was not modified
	content, err := os.ReadFile(cuePath)
	if err != nil {
		t.Fatalf("Failed to read original file: %v", err)
	}
	if !strings.Contains(string(content), "test.wav") {
		t.Error("Original file should not be modified in dry-run mode")
	}
}

// TestValidateCueFile tests CUE file validation
func TestValidateCueFile(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name          string
		content       string
		createAudio   bool
		expectIssues  bool
		issueContains string
	}{
		{
			name: "Valid CUE file",
			content: `FILE "test.flac" WAVE
  TRACK 01 AUDIO
    INDEX 01 00:00:00
`,
			createAudio:  true,
			expectIssues: false,
		},
		{
			name:          "Empty file",
			content:       "",
			createAudio:   false,
			expectIssues:  true,
			issueContains: "empty",
		},
		{
			name: "Missing FILE entry",
			content: `TRACK 01 AUDIO
  INDEX 01 00:00:00
`,
			createAudio:   false,
			expectIssues:  true,
			issueContains: "Missing FILE entry",
		},
		{
			name: "Missing TRACK entry",
			content: `FILE "test.flac" WAVE
  INDEX 01 00:00:00
`,
			createAudio:   false,
			expectIssues:  true,
			issueContains: "Missing TRACK entry",
		},
		{
			name: "No audio files",
			content: `FILE "test.flac" WAVE
  TRACK 01 AUDIO
    INDEX 01 00:00:00
`,
			createAudio:   false,
			expectIssues:  true,
			issueContains: "No audio files",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cuePath := filepath.Join(tmpDir, tt.name+".cue")
			if err := os.WriteFile(cuePath, []byte(tt.content), 0644); err != nil {
				t.Fatalf("Failed to create test CUE file: %v", err)
			}

			if tt.createAudio {
				audioPath := filepath.Join(tmpDir, "test.flac")
				if err := os.WriteFile(audioPath, []byte("dummy"), 0644); err != nil {
					t.Fatalf("Failed to create audio file: %v", err)
				}
				defer os.Remove(audioPath)
			}

			issues := validateCueFile(cuePath)

			if tt.expectIssues && len(issues) == 0 {
				t.Error("Expected issues but got none")
			}
			if !tt.expectIssues && len(issues) > 0 {
				t.Errorf("Expected no issues but got: %v", issues)
			}

			if tt.issueContains != "" {
				found := false
				for _, issue := range issues {
					if strings.Contains(issue, tt.issueContains) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected issue containing %q but got: %v", tt.issueContains, issues)
				}
			}
		})
	}
}

// TestScanAudioFiles tests audio file scanning
func TestScanAudioFiles(t *testing.T) {
	tmpDir := t.TempDir()

	// Create various audio files
	audioFiles := []string{
		"01.flac",
		"02.mp3",
		"03.wav",
		"cover.jpg", // Should be ignored
		"readme.txt", // Should be ignored
	}

	for _, f := range audioFiles {
		path := filepath.Join(tmpDir, f)
		if err := os.WriteFile(path, []byte("dummy"), 0644); err != nil {
			t.Fatalf("Failed to create file %s: %v", f, err)
		}
	}

	found, err := scanAudioFiles(tmpDir)
	if err != nil {
		t.Fatalf("scanAudioFiles failed: %v", err)
	}

	if len(found) != 3 {
		t.Errorf("Expected 3 audio files, got %d", len(found))
	}

	// Check that non-audio files were excluded
	for _, f := range found {
		if strings.HasSuffix(f, ".jpg") || strings.HasSuffix(f, ".txt") {
			t.Errorf("Non-audio file included: %s", f)
		}
	}
}

// TestExtractTrackNumber tests track number extraction
func TestExtractTrackNumber(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"01 - Song.flac", "01"},
		{"02-Song.flac", "02"},
		{"03_Song.flac", "03"},
		{"10.Song.flac", "10"},
		{"Song.flac", ""},
		{"100 Track.flac", "100"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := extractTrackNumber(tt.input)
			if result != tt.expected {
				t.Errorf("extractTrackNumber(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
