package main

import (
	"flag"
	"fmt"
	"os"
)

var (
	recursive   = flag.Bool("r", false, "Recursively process all CUE files in directory")
	dryRun      = flag.Bool("d", false, "Dry-run mode: show changes without writing files")
	verbose     = flag.Bool("v", false, "Verbose output")
	fixMojibake = flag.Bool("m", false, "Fix mojibake (UTF-8 misread as CP1251) in text fields")
	checkMode   = flag.Bool("c", false, "Check mode: validate CUE files and output bash cleanup script for malformed files")
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] <cuefile|directory> [output]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Normalizes CUE file(s) by:\n")
		fmt.Fprintf(os.Stderr, "  - Fixing FILE paths to match actual files in directory\n")
		fmt.Fprintf(os.Stderr, "  - Removing directory prefixes from FILE paths\n")
		fmt.Fprintf(os.Stderr, "  - Fixing file extensions (e.g., .wav -> .flac)\n")
		fmt.Fprintf(os.Stderr, "  - Converting from DOS/Windows encoding to UTF-8\n")
		fmt.Fprintf(os.Stderr, "  - Fixing mojibake (with -m flag) in PERFORMER/TITLE fields\n")
		fmt.Fprintf(os.Stderr, "  - Validating and detecting malformed files (with -c flag)\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  %s album.cue                    # Normalize single file\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -d album.cue                 # Dry-run (show changes only)\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -r /music                    # Recursively process directory\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -r -d /music                 # Recursive dry-run\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -r -c /music > cleanup.sh    # Generate cleanup script for bad files\n", os.Args[0])
	}

	flag.Parse()

	if flag.NArg() < 1 {
		flag.Usage()
		os.Exit(1)
	}

	inputPath := flag.Arg(0)
	outputPath := flag.Arg(1)

	// Check if input is a directory or file
	info, err := os.Stat(inputPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if info.IsDir() {
		// Process directory
		if outputPath != "" {
			fmt.Fprintf(os.Stderr, "Error: Output path cannot be specified when processing a directory\n")
			os.Exit(1)
		}
		if *checkMode {
			checkDirectory(inputPath, *recursive)
		} else {
			processDirectory(inputPath, *recursive, *dryRun, *verbose, *fixMojibake)
		}
	} else {
		// Process single file
		if *recursive {
			fmt.Fprintf(os.Stderr, "Warning: -r flag ignored for single file\n")
		}
		if *checkMode {
			// Check mode for single file
			if issues := validateCueFile(inputPath); len(issues) > 0 {
				fmt.Fprintf(os.Stderr, "# Validation issues found in: %s\n", inputPath)
				for _, issue := range issues {
					fmt.Fprintf(os.Stderr, "#   - %s\n", issue)
				}
				fmt.Printf("rm \"%s\"\n", inputPath)
			} else {
				fmt.Fprintf(os.Stderr, "# File is valid: %s\n", inputPath)
			}
		} else {
			processCueFile(inputPath, outputPath, *dryRun, *verbose, *fixMojibake)
		}
	}
}
