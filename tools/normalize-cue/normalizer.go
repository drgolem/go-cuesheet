package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/drgolem/go-cuesheet/cuesheet/encoding"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
)

// processDirectory processes all CUE files in a directory
func processDirectory(dir string, recursive, dryRun, verbose, fixMojibake bool) {
	var cueFiles []string

	if recursive {
		// Walk directory recursively
		err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() && strings.ToLower(filepath.Ext(path)) == ".cue" {
				cueFiles = append(cueFiles, path)
			}
			return nil
		})
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error walking directory: %v\n", err)
			os.Exit(1)
		}
	} else {
		// Only process files in the specified directory (non-recursive)
		entries, err := os.ReadDir(dir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading directory: %v\n", err)
			os.Exit(1)
		}
		for _, entry := range entries {
			if !entry.IsDir() && strings.ToLower(filepath.Ext(entry.Name())) == ".cue" {
				cueFiles = append(cueFiles, filepath.Join(dir, entry.Name()))
			}
		}
	}

	if len(cueFiles) == 0 {
		fmt.Printf("No CUE files found in %s\n", dir)
		return
	}

	fmt.Printf("Found %d CUE file(s) to process\n\n", len(cueFiles))

	totalProcessed := 0
	totalChanges := 0

	for i, cueFile := range cueFiles {
		fmt.Printf("[%d/%d] Processing: %s\n", i+1, len(cueFiles), cueFile)
		changes := processCueFile(cueFile, "", dryRun, verbose, fixMojibake)
		if changes > 0 {
			totalChanges += changes
			totalProcessed++
		}
		fmt.Println()
	}

	fmt.Printf("Summary: Processed %d file(s) with changes, total %d change(s)\n", totalProcessed, totalChanges)
}

// processCueFile processes a single CUE file
func processCueFile(cuePath, outputPath string, dryRun, verbose, fixMojibake bool) int {
	// If no output path specified, we'll backup original and replace it
	replaceOriginal := (outputPath == "")
	if outputPath == "" {
		outputPath = cuePath
	}

	// Get directory containing the CUE file
	cueDir := filepath.Dir(cuePath)
	if cueDir == "" || cueDir == "." {
		var err error
		cueDir, err = os.Getwd()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting working directory: %v\n", err)
			return 0
		}
	}

	// Scan directory for audio files
	audioFiles, err := scanAudioFiles(cueDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error scanning directory: %v\n", err)
		return 0
	}

	if len(audioFiles) == 0 {
		if verbose {
			fmt.Printf("  Warning: No audio files found in directory %s\n", cueDir)
		}
	} else if verbose {
		fmt.Printf("  Found %d audio file(s) in directory\n", len(audioFiles))
	}

	// Read and normalize CUE file
	lines, err := readCueFile(cuePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading CUE file: %v\n", err)
		if verbose {
			// Show first bytes for debugging encoding issues
			showFileHead(cuePath)
		}
		return 0
	}

	// Normalize FILE lines and optionally fix mojibake
	normalized, changes := normalizeCueLines(lines, audioFiles, verbose, fixMojibake)

	if changes == 0 {
		if verbose {
			fmt.Println("  No changes needed - CUE file is already normalized")
		}
		return 0
	}

	if dryRun {
		// Dry-run mode: print the normalized content
		fmt.Printf("  [DRY-RUN] Would make %d change(s)\n", changes)
		if verbose {
			fmt.Println("  Preview of normalized content:")
			fmt.Println("  " + strings.Repeat("-", 60))
			for _, line := range normalized {
				fmt.Println("  " + line)
			}
			fmt.Println("  " + strings.Repeat("-", 60))
		}
	} else {
		// Backup original file if replacing it
		if replaceOriginal {
			backupPath := cuePath + ".bak"
			if err := os.Rename(cuePath, backupPath); err != nil {
				fmt.Fprintf(os.Stderr, "Error creating backup: %v\n", err)
				return 0
			}
			if verbose {
				fmt.Printf("  ✓ Created backup: %s\n", backupPath)
			}
		}

		// Write normalized CUE file
		if err := writeCueFile(outputPath, normalized); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing normalized CUE file: %v\n", err)
			// Try to restore backup if we renamed the original
			if replaceOriginal {
				backupPath := cuePath + ".bak"
				os.Rename(backupPath, cuePath) // Best effort restore
			}
			return 0
		}

		if replaceOriginal {
			fmt.Printf("  ✓ Normalized CUE file (original saved as %s.bak) - %d change(s)\n", filepath.Base(cuePath), changes)
		} else {
			fmt.Printf("  ✓ Normalized CUE file written to: %s (%d change(s))\n", outputPath, changes)
		}
	}

	return changes
}

// readCueFile reads a CUE file and handles encoding conversion
func readCueFile(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Try reading as UTF-8 first
	scanner := bufio.NewScanner(file)
	var lines []string
	var scanErr error

	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	scanErr = scanner.Err()

	// If UTF-8 failed, try Windows-1252 (common for DOS-format CUE files)
	if scanErr != nil || containsInvalidUTF8(lines) {
		file.Seek(0, 0)
		decoder := charmap.Windows1252.NewDecoder()
		reader := transform.NewReader(file, decoder)
		scanner = bufio.NewScanner(reader)
		lines = lines[:0] // Clear previous attempt

		for scanner.Scan() {
			lines = append(lines, scanner.Text())
		}
		if err := scanner.Err(); err != nil {
			return nil, err
		}
	}

	// Strip UTF-8 BOM from first line if present
	if len(lines) > 0 && strings.HasPrefix(lines[0], "\uFEFF") {
		lines[0] = strings.TrimPrefix(lines[0], "\uFEFF")
	}

	return lines, nil
}

// containsInvalidUTF8 checks if lines contain invalid UTF-8 sequences
func containsInvalidUTF8(lines []string) bool {
	for _, line := range lines {
		for _, r := range line {
			if r == '\uFFFD' { // Unicode replacement character
				return true
			}
		}
	}
	return false
}

// normalizeCueLines normalizes FILE lines and optionally fixes mojibake in CUE content
func normalizeCueLines(lines []string, audioFiles []string, verbose, fixMojibake bool) ([]string, int) {
	// Create a map for faster lookups
	audioMap := make(map[string]string)
	for _, f := range audioFiles {
		// Map by lowercase for case-insensitive matching
		audioMap[strings.ToLower(f)] = f

		// Also map by basename without extension
		base := strings.TrimSuffix(f, filepath.Ext(f))
		audioMap[strings.ToLower(base)] = f
	}

	var normalized []string
	changes := 0
	fileLineRegex := regexp.MustCompile(`^(\s*FILE\s+)"?([^"]+?)"?\s+(WAVE|MP3|AIFF|BINARY|MOTOROLA)?\s*$`)
	textFieldRegex := regexp.MustCompile(`^(\s*(?:PERFORMER|TITLE|SONGWRITER|COMPOSER|ARRANGER|MESSAGE)\s+)"?([^"]+?)"?\s*$`)

	for _, line := range lines {
		// Check if this is a text field line that might need mojibake fixing
		if fixMojibake {
			textMatches := textFieldRegex.FindStringSubmatch(line)
			if textMatches != nil {
				prefix := textMatches[1]
				text := textMatches[2]

				// Try to fix mojibake
				if decoded := encoding.DecodeMojibakeFromCP1251(text); decoded != text {
					if verbose {
						fmt.Printf("  ✓ Fixed mojibake: %s -> %s\n", text, decoded)
					}
					newLine := fmt.Sprintf("%s\"%s\"", prefix, decoded)
					normalized = append(normalized, newLine)
					changes++
					continue
				}
			}
		}

		// Check if this is a FILE line
		matches := fileLineRegex.FindStringSubmatch(line)
		if matches == nil {
			normalized = append(normalized, line)
			continue
		}

		prefix := matches[1]   // "  FILE "
		filePath := matches[2] // The file path
		fileType := matches[3] // WAVE, MP3, etc.
		if fileType == "" {
			fileType = "WAVE"
		}

		// Extract just the filename (remove directory path)
		// Handle both Unix (/) and Windows (\) path separators
		// Convert backslashes to forward slashes for filepath.Base to work
		normalizedPath := strings.ReplaceAll(filePath, "\\", "/")
		fileName := filepath.Base(normalizedPath)

		// Try to find matching audio file
		matchedFile := findMatchingAudioFile(fileName, audioMap)

		if matchedFile != "" && matchedFile != fileName {
			if verbose {
				fmt.Printf("  ✓ Fixed: %s -> %s\n", fileName, matchedFile)
			}
			fileName = matchedFile
			changes++
		} else if matchedFile == "" && len(audioFiles) > 0 {
			// No match found, but we have audio files
			if verbose {
				fmt.Printf("  ⚠ Warning: No matching file found for: %s\n", fileName)
			}
		}

		// Reconstruct the FILE line with proper quoting
		newLine := fmt.Sprintf("%s\"%s\" %s", prefix, fileName, fileType)
		normalized = append(normalized, newLine)
	}

	return normalized, changes
}

// writeCueFile writes normalized CUE content to file as UTF-8
func writeCueFile(path string, lines []string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	for _, line := range lines {
		if _, err := writer.WriteString(line + "\n"); err != nil {
			return err
		}
	}
	return writer.Flush()
}

// showFileHead displays first bytes of file for debugging
func showFileHead(path string) {
	file, err := os.Open(path)
	if err != nil {
		return
	}
	defer file.Close()

	buf := make([]byte, 32)
	n, err := file.Read(buf)
	if err != nil && n == 0 {
		return
	}

	fmt.Fprintf(os.Stderr, "  First %d bytes (hex): ", n)
	for i := 0; i < n && i < 16; i++ {
		fmt.Fprintf(os.Stderr, "%02x ", buf[i])
	}
	fmt.Fprintf(os.Stderr, "\n")
}
