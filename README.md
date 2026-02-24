# go-cuesheet

## Description

Cuesheet (.cue) reader/writer library written in Go.

Forked from github.com/mtojo/go-cuesheet/

## Installation

    go get github.com/drgolem/go-cuesheet/cuesheet

## Examples

See the [examples](examples/) directory for usage examples:

- [print-tracks](examples/print-tracks/) - Parse a CUE file and display track information in tabular format

## Tools

See the [tools](tools/) directory for utility tools:

- [normalize-cue](tools/normalize-cue/) - Fix FILE paths and extensions to match actual audio files, convert encoding to UTF-8, fix mojibake in metadata fields
- [decode-mojibake](tools/decode-mojibake/) - Decode garbled Cyrillic text from the command line

## Dependencies

- [cyrillic-encoding](https://github.com/drgolem/cyrillic-encoding) - Cyrillic encoding detection and mojibake fixing
- [golang.org/x/text](https://pkg.go.dev/golang.org/x/text) - Character encoding support

## License

MIT
