# normalize-cue

A tool to normalize CUE sheet files by fixing FILE paths and extensions to match actual audio files in the directory.

## Features

- **Fixes FILE paths**: Removes directory prefixes from FILE entries
- **Corrects extensions**: Matches FILE entries to actual files (e.g., .wav → .flac)
- **Encoding conversion**: Converts DOS/Windows-1252 encoded CUE files to UTF-8
- **Mojibake fixing**: Fixes double-encoded Cyrillic text (UTF-8 misread as CP1251)
- **Validation mode**: Detects empty or malformed CUE files and generates cleanup scripts
- **Smart matching**: Matches files by name, basename, or track number
- **Safe replacement**: Backs up original as .bak before replacing with normalized version

## Common Issues It Fixes

### 1. Directory Prefixes
```
Before: FILE "Various - Journey Through The Past\04 - Ohio.wav" WAVE
After:  FILE "04 - Ohio.flac" WAVE
```

### 2. Wrong Extensions
```
Before: FILE "01 - Song.wav" WAVE
After:  FILE "01 - Song.flac" WAVE  (when actual file is .flac)
```

### 3. Encoding Issues
- Converts Windows-1252 / DOS format to UTF-8
- Strips UTF-8 BOM (Byte Order Mark) if present
- Fixes special characters and accents

### 4. Mojibake (Double-Encoded Text)
```
Before: PERFORMER "Р'СЂР°РІРѕ"
After:  PERFORMER "Браво"  (with -m flag)
```

### 5. Malformed Files
- Validates CUE files for required fields (FILE, TRACK, INDEX)
- Detects empty or corrupt files
- Checks if audio files exist in directory

## Installation

```bash
cd tools/normalize-cue
go build
```

## Usage

```bash
# Single file - backs up original as input.cue.bak and replaces with normalized version
./normalize-cue input.cue

# Specify output file (keeps original unchanged)
./normalize-cue input.cue output.cue

# Dry-run mode (show changes without writing)
./normalize-cue -d input.cue

# Dry-run with verbose output (shows preview)
./normalize-cue -d -v input.cue

# Process all CUE files in directory (non-recursive)
./normalize-cue directory/

# Recursively process all CUE files in directory tree
./normalize-cue -r directory/

# Recursive dry-run
./normalize-cue -r -d /path/to/music

# Fix mojibake (double-encoded Cyrillic text)
./normalize-cue -m russian-album.cue

# Check mode - validate CUE files and generate cleanup script
./normalize-cue -c directory/
./normalize-cue -r -c /path/to/music > cleanup.sh
```

## Options

- `-d` - Dry-run mode: show what would be changed without writing files
- `-r` - Recursively process all CUE files in directory and subdirectories
- `-v` - Verbose output: show detailed changes and preview
- `-m` - Fix mojibake (UTF-8 text misread as CP1251) in PERFORMER/TITLE fields
- `-c` - Check mode: validate CUE files and output bash cleanup script for malformed files

## Examples

### Single File Example

Given a directory with these files:
```
01 - For What's It Worth,Mr. Soul.flac
02 - Rock & Roll Woman.flac
03 - Find The Cost Of Freedom.flac
04 - Ohio.flac
05 - Southern Man.flac
```

And a CUE file with wrong paths:
```
FILE "Various - Journey Through The Past\01 - For What's It Worth,Mr. Soul.wav" WAVE
  TRACK 01 AUDIO
    TITLE "For What's It Worth/Mr. Soul"
    INDEX 01 00:00:00
FILE "Various - Journey Through The Past\02 - Rock & Roll Woman.wav" WAVE
  TRACK 02 AUDIO
    TITLE "Rock & Roll Woman"
    INDEX 01 00:00:00
```

Run the tool:
```bash
$ ./normalize-cue album.cue

Found 5 audio file(s) in directory:
  - 01 - For What's It Worth,Mr. Soul.flac
  - 02 - Rock & Roll Woman.flac
  - 03 - Find The Cost Of Freedom.flac
  - 04 - Ohio.flac
  - 05 - Southern Man.flac

✓ Fixed: 01 - For What's It Worth,Mr. Soul.wav
      -> 01 - For What's It Worth,Mr. Soul.flac
✓ Fixed: 02 - Rock & Roll Woman.wav
      -> 02 - Rock & Roll Woman.flac
✓ Fixed: 03 - Find The Cost Of Freedom.wav
      -> 03 - Find The Cost Of Freedom.flac
✓ Fixed: 04 - Ohio.wav
      -> 04 - Ohio.flac
✓ Fixed: 05 - Southern Man.wav
      -> 05 - Southern Man.flac

✓ Normalized CUE file (original saved as album.cue.bak) - 5 change(s)
```

### Dry-Run Example

Check what would be changed without modifying files:

```bash
$ ./normalize-cue -d -v album.cue

  Found 6 audio file(s) in directory
  ✓ Fixed: 01 - Song.wav -> 01 - Song.flac
  ✓ Fixed: 02 - Song.wav -> 02 - Song.flac
  [DRY-RUN] Would make 2 change(s)
  Preview of normalized content:
  ------------------------------------------------------------
  REM GENRE Rock
  FILE "01 - Song.flac" WAVE
    TRACK 01 AUDIO
      TITLE "First Song"
  ...
  ------------------------------------------------------------
```

### Recursive Directory Example

Process an entire music library:

```bash
$ ./normalize-cue -r -d /Music

Found 15 CUE file(s) to process

[1/15] Processing: /Music/Artist1/Album1/album.cue
  [DRY-RUN] Would make 3 change(s)

[2/15] Processing: /Music/Artist1/Album2/album.cue
  No changes needed

[3/15] Processing: /Music/Artist2/BestOf/tracks.cue
  [DRY-RUN] Would make 12 change(s)
...

Summary: Processed 8 file(s) with changes, total 45 change(s)
```

### Check Mode - Validate and Generate Cleanup Script

Detect malformed CUE files and generate a bash script for cleanup:

```bash
$ ./normalize-cue -r -c /Music > cleanup.sh

#!/bin/bash
# CUE file cleanup script - generated by normalize-cue
# Found 25 CUE file(s) to validate
# Review this script before executing!

# [MALFORMED] /Music/Broken/empty.cue
#   - File is empty (0 bytes)
rm "/Music/Broken/empty.cue"

# [MALFORMED] /Music/Broken/missing-fields.cue
#   - Missing FILE entry
#   - Missing TRACK entry
#   - Missing INDEX entry
rm "/Music/Broken/missing-fields.cue"

# [MALFORMED] /Music/Broken/no-audio.cue
#   - No FILE entries match actual audio files in directory
rm "/Music/Broken/no-audio.cue"

# Found 3 malformed file(s) out of 25 total
```

Review the generated script, then execute it:
```bash
bash cleanup.sh
```

**Validation Checks:**
- Empty files (0 bytes)
- Files with only whitespace
- Missing required fields (FILE, TRACK, INDEX)
- No matching audio files in directory
- Unparseable/corrupt files

### Mojibake Fixing

Fix double-encoded Cyrillic text (common in Russian CUE files):

```bash
$ ./normalize-cue -m -v russian-album.cue

  Found 4 audio file(s) in directory
  ✓ Fixed mojibake: Р'СЂР°РІРѕ -> Браво
  ✓ Fixed mojibake: РЎС‚РёР»СЏРіРё РёР· РњРѕСЃРєРІС‹ -> Стиляги из Москвы
  ✓ Normalized CUE file (original saved as russian-album.cue.bak) - 2 change(s)
```

## How It Works

1. **Scans directory** for audio files (.flac, .wav, .mp3, .ape, .wv, .m4a, .ogg, .opus, .aiff)
2. **Reads CUE file** with encoding detection (UTF-8 or Windows-1252)
3. **Matches FILE entries** to actual files using:
   - Exact filename match (case-insensitive)
   - Basename match (without extension)
   - Track number prefix match (e.g., "01 -", "02-")
4. **Updates FILE lines** with correct filenames and extensions
5. **Writes normalized CUE** as UTF-8

## Supported Audio Formats

- FLAC (.flac)
- WAV (.wav)
- MP3 (.mp3)
- APE (.ape)
- WavPack (.wv)
- M4A (.m4a)
- OGG (.ogg)
- OPUS (.opus)
- AIFF (.aiff, .aif)

## Notes

- The tool preserves all CUE metadata (REM comments, titles, performers, etc.)
- Only FILE lines and text fields (with -m flag) are modified
- Original CUE file is backed up as .bak before being replaced
- When specifying an output file, the original is kept unchanged
- If no matches are found, the tool will warn but still create output
