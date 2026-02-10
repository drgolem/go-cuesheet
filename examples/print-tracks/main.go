package main

import (
	"fmt"
	"os"

	"github.com/drgolem/go-cuesheet/cuesheet"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <cuefile>\n", os.Args[0])
		os.Exit(1)
	}

	// Open the CUE file
	file, err := os.Open(os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening file: %v\n", err)
		os.Exit(1)
	}
	defer file.Close()

	// Parse the CUE file
	cs, err := cuesheet.ReadFile(file)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing CUE file: %v\n", err)
		os.Exit(1)
	}

	// Print album information
	if cs.Title != "" {
		fmt.Printf("Album: %s\n", cs.Title)
	}
	if cs.Performer != "" {
		fmt.Printf("Artist: %s\n", cs.Performer)
	}
	fmt.Println()

	// Print header
	fmt.Println("Track | Title                          | Performer                      | Duration")
	fmt.Println("------|--------------------------------|--------------------------------|----------")

	// Iterate through all tracks
	for i := range cs.File {
		for j := range cs.File[i].Tracks {
			track := &cs.File[i].Tracks[j]

			// Calculate duration
			var duration string
			if j+1 < len(cs.File[i].Tracks) {
				// Not the last track in this file
				nextTrack := &cs.File[i].Tracks[j+1]
				if len(nextTrack.Index) > 0 {
					dur := track.Duration(nextTrack.Index[0].Frame)
					minutes := int(dur.Minutes())
					seconds := int(dur.Seconds()) % 60
					duration = fmt.Sprintf("%02d:%02d", minutes, seconds)
				} else {
					duration = "unknown"
				}
			} else if i+1 < len(cs.File) {
				// Last track in this file, check if there's another file
				if len(cs.File[i+1].Tracks) > 0 {
					nextTrack := &cs.File[i+1].Tracks[0]
					if len(nextTrack.Index) > 0 {
						dur := track.Duration(nextTrack.Index[0].Frame)
						minutes := int(dur.Minutes())
						seconds := int(dur.Seconds()) % 60
						duration = fmt.Sprintf("%02d:%02d", minutes, seconds)
					} else {
						duration = "unknown"
					}
				} else {
					duration = "unknown"
				}
			} else {
				// Last track overall
				duration = "unknown"
			}

			// Print track information in columnar format
			title := track.Title
			if title == "" {
				title = "-"
			}

			// Use track performer, fall back to album performer
			performer := track.Performer
			if performer == "" {
				performer = cs.Performer
			}
			if performer == "" {
				performer = "-"
			}

			fmt.Printf("%5d | %-30s | %-30s | %s\n",
				track.TrackNumber,
				truncate(title, 30),
				truncate(performer, 30),
				duration)
		}
	}

	fmt.Printf("\nTotal tracks: %d\n", cs.TrackCount())
}

// truncate truncates a string to the specified length
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
