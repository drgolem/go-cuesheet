package main

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// AudioExtensions lists common audio file extensions
var AudioExtensions = map[string]bool{
	".flac": true,
	".wav":  true,
	".mp3":  true,
	".ape":  true,
	".wv":   true,
	".m4a":  true,
	".ogg":  true,
	".opus": true,
	".aiff": true,
	".aif":  true,
}

// scanAudioFiles scans a directory for audio files
func scanAudioFiles(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var audioFiles []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		ext := strings.ToLower(filepath.Ext(entry.Name()))
		if AudioExtensions[ext] {
			audioFiles = append(audioFiles, entry.Name())
		}
	}

	// Sort for consistent ordering
	sort.Strings(audioFiles)
	return audioFiles, nil
}

// findMatchingAudioFile finds the best matching audio file
func findMatchingAudioFile(fileName string, audioMap map[string]string) string {
	// Direct match (case-insensitive)
	if match, ok := audioMap[strings.ToLower(fileName)]; ok {
		return match
	}

	// Try without extension
	base := strings.TrimSuffix(fileName, filepath.Ext(fileName))
	if match, ok := audioMap[strings.ToLower(base)]; ok {
		return match
	}

	// Try extracting track number and matching by that
	trackNum := extractTrackNumber(fileName)
	if trackNum != "" {
		for audioFile := range audioMap {
			if strings.HasPrefix(audioFile, trackNum+" ") ||
				strings.HasPrefix(audioFile, trackNum+"-") ||
				strings.HasPrefix(audioFile, trackNum+"_") {
				return audioMap[audioFile]
			}
		}
	}

	// No match found
	return ""
}

// extractTrackNumber extracts track number from filename (e.g., "01", "02")
func extractTrackNumber(fileName string) string {
	re := regexp.MustCompile(`^(\d{1,3})[\s\-_\.]`)
	matches := re.FindStringSubmatch(fileName)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}
