# Print Tracks Example

A simple command-line tool that parses a CUE sheet file and displays track information in a tabular format.

## Usage

```bash
go run main.go <cuefile>
```

Or build and run:

```bash
go build
./print-tracks <cuefile>
```

## Example

```bash
$ go run main.go ../../cuesheet/testdata/sample_1.cue

Album: Album Title
Artist: Artist Name

Track | Title                          | Performer                      | Duration
------|--------------------------------|--------------------------------|----------
    1 | First Song                     | Artist Name                    | 05:30
    2 | Second Song                    | Artist Name                    | 04:45
    3 | Third Song                     | Artist Name                    | unknown

Total tracks: 3
```

## Output Format

The program displays:
- Album title and artist (if available)
- Table with columns:
  - **Track**: Track number
  - **Title**: Track title
  - **Performer**: Track performer (or "-" if not specified)
  - **Duration**: Track duration in MM:SS format

## Notes

- Duration is calculated from the difference between track start positions
- For the last track or multi-file CUE sheets where each track has its own file, duration is shown as "unknown" (would require reading actual audio files)
- Long titles/performers are truncated to 30 characters with "..."
