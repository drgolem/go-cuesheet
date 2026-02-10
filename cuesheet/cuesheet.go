package cuesheet

import (
	"bufio"
	"errors"
	"io"
	"strconv"
	"strings"
	"time"
)

const (
	delims          = "\t\n\r "
	eol             = "\n"
	framesPerSecond = 75
)

// Frame represents CD audio time in frames
// The CD standard uses 75 frames per second
// Time format: MSF (Minutes:Seconds:Frames) e.g., "03:45:22" = 3 minutes, 45 seconds, 22 frames
type Frame uint64

type Flags int

// RemType represents the type of a structured REM comment
type RemType int

const (
	RemUnknown RemType = iota
	RemDate
	RemGenre
	RemDiscNumber
	RemComment
	RemReplayGainAlbumGain
	RemReplayGainAlbumPeak
	RemReplayGainTrackGain
	RemReplayGainTrackPeak
)

// RemField represents a parsed REM comment field
type RemField struct {
	Type  RemType
	Key   string
	Value string
}

const (
	None    Flags = iota
	Dcp           = 1 << iota
	Four_ch       = 1 << iota
	Pre           = 1 << iota
	Scms          = 1 << iota
)

// TrackIndex represents a track index position
// INDEX 00: Pregap start - marks the beginning of the pregap (audio/silence before track starts)
// INDEX 01: Track start - marks the actual beginning of the track (required for all tracks)
// INDEX 02+: Sub-indexes within the track (optional)
//
// Example:
//   INDEX 00 03:00:00  - Pregap starts at 3 minutes
//   INDEX 01 03:02:00  - Track starts at 3:02 (with 2 second pregap)
type TrackIndex struct {
	Number uint  // Index number (0-99, where 0=pregap, 1=track start)
	Frame  Frame // Position in MSF time format
}

type Track struct {
	TrackNumber   uint
	TrackDataType string
	Flags         Flags
	Isrc          string
	Title         string
	Performer     string
	SongWriter    string
	Composer      string // CD-TEXT: track composer
	Arranger      string // CD-TEXT: track arranger
	Message       string // CD-TEXT: track message
	Pregap        Frame
	Postgap       Frame
	Index         []TrackIndex
}

type File struct {
	FileName string
	FileType string
	Tracks   []Track
}

type Cuesheet struct {
	Rem        []string
	Catalog    string
	CdTextFile string
	Title      string
	Performer  string
	SongWriter string
	Composer   string // CD-TEXT: album composer
	Arranger   string // CD-TEXT: album arranger
	Message    string // CD-TEXT: album message
	Genre      string // CD-TEXT: album genre
	DiscId     string // CD-TEXT: disc ID
	UpcEan     string // CD-TEXT: UPC/EAN barcode
	Pregap     Frame
	Postgap    Frame
	File       []File
}

func ReadFile(r io.Reader) (*Cuesheet, error) {
	b := bufio.NewReader(r)
	cuesheet := &Cuesheet{}

	for {
		line, err := (*b).ReadString('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		line = strings.Trim(line, delims)
		command := ReadString(&line)

		switch command {
		case "REM":
			cuesheet.Rem = append(cuesheet.Rem, line)
		case "CATALOG":
			cuesheet.Catalog = line
		case "CDTEXTFILE":
			cuesheet.CdTextFile = ReadString(&line)
		case "TITLE":
			cuesheet.Title = ReadString(&line)
		case "PERFORMER":
			cuesheet.Performer = ReadString(&line)
		case "SONGWRITER":
			cuesheet.SongWriter = ReadString(&line)
		case "COMPOSER":
			cuesheet.Composer = ReadString(&line)
		case "ARRANGER":
			cuesheet.Arranger = ReadString(&line)
		case "MESSAGE":
			cuesheet.Message = ReadString(&line)
		case "GENRE":
			cuesheet.Genre = ReadString(&line)
		case "DISC_ID":
			cuesheet.DiscId = ReadString(&line)
		case "UPC_EAN":
			cuesheet.UpcEan = ReadString(&line)
		case "PREGAP":
			frame, err := ReadFrame(&line)
			if err != nil {
				return nil, err
			}
			cuesheet.Pregap = frame
		case "POSTGAP":
			frame, err := ReadFrame(&line)
			if err != nil {
				return nil, err
			}
			cuesheet.Postgap = frame
		case "FILE":
			fname := ReadString(&line)
			ftype := ReadString(&line)
			tracks, err := readTracks(b)
			if err != nil {
				return nil, err
			}
			cuesheet.File = append(cuesheet.File, File{fname, ftype, *tracks})
		}
	}

	return cuesheet, nil
}

func WriteFile(w io.Writer, cuesheet *Cuesheet) error {
	ws := bufio.NewWriter(w)

	for i := 0; i < len(cuesheet.Rem); i++ {
		ws.WriteString("REM " + cuesheet.Rem[i] + eol)
	}

	if len(cuesheet.Catalog) > 0 {
		ws.WriteString("CATALOG " + cuesheet.Catalog + eol)
	}

	if len(cuesheet.CdTextFile) > 0 {
		ws.WriteString("CDTEXTFILE " + FormatString(cuesheet.CdTextFile) + eol)
	}

	if len(cuesheet.Title) > 0 {
		ws.WriteString("TITLE " + FormatString(cuesheet.Title) + eol)
	}

	if len(cuesheet.Performer) > 0 {
		ws.WriteString("PERFORMER " + FormatString(cuesheet.Performer) + eol)
	}

	if len(cuesheet.SongWriter) > 0 {
		ws.WriteString("SONGWRITER " + FormatString(cuesheet.SongWriter) + eol)
	}

	if len(cuesheet.Composer) > 0 {
		ws.WriteString("COMPOSER " + FormatString(cuesheet.Composer) + eol)
	}

	if len(cuesheet.Arranger) > 0 {
		ws.WriteString("ARRANGER " + FormatString(cuesheet.Arranger) + eol)
	}

	if len(cuesheet.Message) > 0 {
		ws.WriteString("MESSAGE " + FormatString(cuesheet.Message) + eol)
	}

	if len(cuesheet.Genre) > 0 {
		ws.WriteString("GENRE " + FormatString(cuesheet.Genre) + eol)
	}

	if len(cuesheet.DiscId) > 0 {
		ws.WriteString("DISC_ID " + FormatString(cuesheet.DiscId) + eol)
	}

	if len(cuesheet.UpcEan) > 0 {
		ws.WriteString("UPC_EAN " + FormatString(cuesheet.UpcEan) + eol)
	}

	if cuesheet.Pregap > 0 {
		ws.WriteString("PREGAP " + FormatFrame(cuesheet.Pregap) + eol)
	}

	if cuesheet.Postgap > 0 {
		ws.WriteString("POSTGAP " + FormatFrame(cuesheet.Postgap) + eol)
	}

	for i := 0; i < len(cuesheet.File); i++ {
		file := cuesheet.File[i]
		ws.WriteString("FILE " + FormatString(file.FileName) +
			" " + file.FileType + eol)

		for i := 0; i < len(file.Tracks); i++ {
			track := file.Tracks[i]

			ws.WriteString("  TRACK " + FormatTrackNumber(track.TrackNumber) +
				" " + track.TrackDataType + eol)

			if track.Flags != None {
				ws.WriteString("    FLAGS")
				if (track.Flags & Dcp) != 0 {
					ws.WriteString(" DCP")
				}
				if (track.Flags & Four_ch) != 0 {
					ws.WriteString(" 4CH")
				}
				if (track.Flags & Pre) != 0 {
					ws.WriteString(" PRE")
				}
				if (track.Flags & Scms) != 0 {
					ws.WriteString(" SCMS")
				}
				ws.WriteString(eol)
			}

			if len(track.Isrc) > 0 {
				ws.WriteString("    ISRC " + track.Isrc + eol)
			}

			if len(track.Title) > 0 {
				ws.WriteString("    TITLE " + FormatString(track.Title) + eol)
			}

			if len(track.Performer) > 0 {
				ws.WriteString("    PERFORMER " + FormatString(track.Performer) + eol)
			}

			if len(track.SongWriter) > 0 {
				ws.WriteString("    SONGWRITER " + FormatString(track.SongWriter) + eol)
			}

			if len(track.Composer) > 0 {
				ws.WriteString("    COMPOSER " + FormatString(track.Composer) + eol)
			}

			if len(track.Arranger) > 0 {
				ws.WriteString("    ARRANGER " + FormatString(track.Arranger) + eol)
			}

			if len(track.Message) > 0 {
				ws.WriteString("    MESSAGE " + FormatString(track.Message) + eol)
			}

			if track.Pregap > 0 {
				ws.WriteString("    PREGAP " + FormatFrame(track.Pregap) + eol)
			}

			if track.Postgap > 0 {
				ws.WriteString("    POSTGAP " + FormatFrame(track.Postgap) + eol)
			}

			for i := 0; i < len(track.Index); i++ {
				index := track.Index[i]
				ws.WriteString("    INDEX " + FormatTrackNumber(index.Number) +
					" " + FormatFrame(index.Frame) + eol)
			}
		}
	}

	ws.Flush()

	return nil
}

func ReadString(s *string) string {
	*s = strings.TrimLeft(*s, delims)
	if isQuoted(*s) {
		v := unquote(*s)
		*s = (*s)[len(v)+2:]
		return v
	}
	for i := 0; i < len(*s); i++ {
		if (*s)[i] == ' ' {
			v := (*s)[0:i]
			*s = (*s)[i+1:]
			return v
		}
	}
	v := *s
	*s = ""
	return v
}

func ReadInt(s *string) (int, error) {
	v := ReadString(s)
	n, err := strconv.ParseInt(v, 10, 32)
	if err != nil {
		return 0, err
	}
	return int(n), nil
}

func ReadUint(s *string) (uint, error) {
	v := ReadString(s)
	n, err := strconv.ParseUint(v, 10, 32)
	if err != nil {
		return 0, err
	}
	return uint(n), nil
}

func ReadFrame(s *string) (Frame, error) {
	v := strings.Split(ReadString(s), ":")
	if len(v) != 3 {
		return 0, strconv.ErrSyntax
	}
	mm, err := strconv.ParseUint(v[0], 10, 32)
	if err != nil {
		return 0, err
	}
	ss, err := strconv.ParseUint(v[1], 10, 32)
	if err != nil {
		return 0, err
	}
	ff, err := strconv.ParseUint(v[2], 10, 32)
	if err != nil {
		return 0, err
	}
	return Frame((mm*60+ss)*framesPerSecond + ff), nil
}

func FormatString(s string) string {
	if strings.ContainsAny(s, delims) {
		return quote(s, '"')
	}
	return s
}

func FormatTrackNumber(n uint) string {
	return leftPad(strconv.FormatUint(uint64(n), 10), "0", 2)
}

func FormatFrame(frame Frame) string {
	n := frame / framesPerSecond
	mm := n / 60
	ss := n % 60
	ff := frame % framesPerSecond
	return leftPad(strconv.FormatUint(uint64(mm), 10), "0", 2) + ":" +
		leftPad(strconv.FormatUint(uint64(ss), 10), "0", 2) + ":" +
		leftPad(strconv.FormatUint(uint64(ff), 10), "0", 2)
}

func isQuoted(s string) bool {
	if s == "" {
		return false
	}
	return s[0] == '"' || s[0] == '\''
}

func quote(s string, quote byte) string {
	buf := make([]byte, 0, 3*len(s)/2)
	buf = append(buf, quote)
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c == quote || c == '\\' {
			buf = append(buf, '\\')
			buf = append(buf, byte(c))
		} else {
			buf = append(buf, byte(c))
		}
	}
	buf = append(buf, quote)
	return string(buf)
}

func unquote(s string) string {
	quote := s[0]
	i := 1
	for ; i < len(s); i++ {
		if s[i] == quote {
			break
		}
		if s[i] == '\\' {
			i++
		}
	}
	return s[1:i]
}

func readTrack(b *bufio.Reader, track *Track) error {
L:
	for {
		before := *b
		line, err := (*b).ReadString('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		if !strings.HasPrefix(line, "    ") {
			*b = before
			break
		}
		line = strings.Trim(line, delims)
		command := ReadString(&line)

		switch command {
		case "FLAGS":
			track.Flags = None
			for len(line) > 0 {
				switch ReadString(&line) {
				case "DCP":
					track.Flags |= Dcp
				case "4CH":
					track.Flags |= Four_ch
				case "PRE":
					track.Flags |= Pre
				case "SCMS":
					track.Flags |= Scms
				}
			}
		case "ISRC":
			track.Isrc = line
		case "TITLE":
			track.Title = ReadString(&line)
		case "PERFORMER":
			track.Performer = ReadString(&line)
		case "SONGWRITER":
			track.SongWriter = ReadString(&line)
		case "COMPOSER":
			track.Composer = ReadString(&line)
		case "ARRANGER":
			track.Arranger = ReadString(&line)
		case "MESSAGE":
			track.Message = ReadString(&line)
		case "PREGAP":
			frame, err := ReadFrame(&line)
			if err != nil {
				return err
			}
			track.Pregap = frame
		case "POSTGAP":
			frame, err := ReadFrame(&line)
			if err != nil {
				return err
			}
			track.Postgap = frame
		case "INDEX":
			index := TrackIndex{}
			num, err := ReadUint(&line)
			if err != nil {
				return err
			}
			index.Number = num
			frame, err := ReadFrame(&line)
			if err != nil {
				return err
			}
			index.Frame = frame
			track.Index = append(track.Index, index)
		case "REM":
			// ignore comment inside of track
		default:
			break L
		}
	}

	return nil
}

func readTracks(b *bufio.Reader) (*[]Track, error) {
	tracks := &[]Track{}

L:
	for {
		before := *b
		line, err := (*b).ReadString('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		if !strings.HasPrefix(line, "  ") {
			*b = before
			break
		}
		line = strings.Trim(line, delims)
		command := ReadString(&line)

		switch command {
		case "TRACK":
			track := Track{}
			num, err := ReadUint(&line)
			if err != nil {
				return nil, err
			}
			track.TrackNumber = num
			track.TrackDataType = ReadString(&line)
			if err := readTrack(b, &track); err != nil {
				return nil, err
			}
			*tracks = append(*tracks, track)
		default:
			break L
		}
	}

	return tracks, nil
}

func leftPad(s, padStr string, overallLen int) string {
	var padCountInt int
	padCountInt = 1 + ((overallLen - len(padStr)) / len(padStr))
	var retStr = strings.Repeat(padStr, padCountInt) + s
	return retStr[(len(retStr) - overallLen):]
}

// REM parsing functions

// ParseRemComment parses a REM comment line into a structured RemField
// Common formats:
//   REM DATE "2024"
//   REM GENRE "Rock"
//   REM DISCNUMBER 1
//   REM COMMENT "Text"
//   REM REPLAYGAIN_ALBUM_GAIN -6.2 dB
func ParseRemComment(rem string) (*RemField, bool) {
	if len(rem) == 0 {
		return nil, false
	}

	// Parse key-value from REM comment
	parts := strings.SplitN(rem, " ", 2)
	if len(parts) < 1 {
		return nil, false
	}

	key := strings.ToUpper(parts[0])
	value := ""
	if len(parts) == 2 {
		value = strings.TrimSpace(parts[1])
		// Remove quotes if present
		if len(value) > 0 && (value[0] == '"' || value[0] == '\'') {
			value = unquote(value)
		}
	}

	field := &RemField{
		Key:   key,
		Value: value,
	}

	// Determine RemType
	switch key {
	case "DATE":
		field.Type = RemDate
	case "GENRE":
		field.Type = RemGenre
	case "DISCNUMBER":
		field.Type = RemDiscNumber
	case "COMMENT":
		field.Type = RemComment
	case "REPLAYGAIN_ALBUM_GAIN":
		field.Type = RemReplayGainAlbumGain
	case "REPLAYGAIN_ALBUM_PEAK":
		field.Type = RemReplayGainAlbumPeak
	case "REPLAYGAIN_TRACK_GAIN":
		field.Type = RemReplayGainTrackGain
	case "REPLAYGAIN_TRACK_PEAK":
		field.Type = RemReplayGainTrackPeak
	default:
		field.Type = RemUnknown
	}

	return field, true
}

// GetRemFields returns all parsed REM fields from the cuesheet
func (c *Cuesheet) GetRemFields() []RemField {
	var fields []RemField
	for _, rem := range c.Rem {
		if field, ok := ParseRemComment(rem); ok {
			fields = append(fields, *field)
		}
	}
	return fields
}

// GetRemValue returns the value of the first REM field with the given type
func (c *Cuesheet) GetRemValue(typ RemType) (string, bool) {
	for _, rem := range c.Rem {
		if field, ok := ParseRemComment(rem); ok && field.Type == typ {
			return field.Value, true
		}
	}
	return "", false
}

// GetRemByKey returns the value of the first REM field with the given key
func (c *Cuesheet) GetRemByKey(key string) (string, bool) {
	upperKey := strings.ToUpper(key)
	for _, rem := range c.Rem {
		if field, ok := ParseRemComment(rem); ok && field.Key == upperKey {
			return field.Value, true
		}
	}
	return "", false
}

// Helper methods

// GetTrack returns the track with the specified number
func (c *Cuesheet) GetTrack(number uint) (*Track, error) {
	for i := range c.File {
		for j := range c.File[i].Tracks {
			if c.File[i].Tracks[j].TrackNumber == number {
				return &c.File[i].Tracks[j], nil
			}
		}
	}
	return nil, errors.New("track not found")
}

// TrackCount returns the total number of tracks across all files
func (c *Cuesheet) TrackCount() int {
	count := 0
	for i := range c.File {
		count += len(c.File[i].Tracks)
	}
	return count
}

// TotalDuration calculates the total duration of all tracks
// Returns the duration from the start of the first track to the end of the last track
func (c *Cuesheet) TotalDuration() time.Duration {
	if len(c.File) == 0 {
		return 0
	}
	var lastFrame Frame
	for i := range c.File {
		for j := range c.File[i].Tracks {
			track := &c.File[i].Tracks[j]
			if len(track.Index) > 0 {
				for k := range track.Index {
					if track.Index[k].Frame > lastFrame {
						lastFrame = track.Index[k].Frame
					}
				}
			}
		}
	}
	return lastFrame.ToDuration()
}

// GetIndex returns the index with the specified number
func (t *Track) GetIndex(number uint) (*TrackIndex, error) {
	for i := range t.Index {
		if t.Index[i].Number == number {
			return &t.Index[i], nil
		}
	}
	return nil, errors.New("index not found")
}

// IndexCount returns the number of indexes in the track
func (t *Track) IndexCount() int {
	return len(t.Index)
}

// StartPosition returns the position of INDEX 01 (the actual track start)
func (t *Track) StartPosition() (Frame, error) {
	idx, err := t.GetIndex(1)
	if err != nil {
		return 0, errors.New("track missing INDEX 01")
	}
	return idx.Frame, nil
}

// HasPregap returns true if the track has INDEX 00 (pregap marker)
func (t *Track) HasPregap() bool {
	_, err := t.GetIndex(0)
	return err == nil
}

// GetPregapIndex returns INDEX 00 if it exists
// INDEX 00 marks the pregap (audio/silence before the track starts)
func (t *Track) GetPregapIndex() (*TrackIndex, bool) {
	idx, err := t.GetIndex(0)
	if err != nil {
		return nil, false
	}
	return idx, true
}

// GetStartIndex returns INDEX 01 (required track start position)
func (t *Track) GetStartIndex() (*TrackIndex, error) {
	return t.GetIndex(1)
}

// PregapDuration calculates the pregap duration
// If INDEX 00 exists, returns the difference between INDEX 00 and INDEX 01
// Otherwise returns the PREGAP field value
func (t *Track) PregapDuration() time.Duration {
	idx00, hasIdx00 := t.GetPregapIndex()
	idx01, err := t.GetStartIndex()
	if err != nil {
		return 0
	}

	if hasIdx00 {
		// Calculate from index difference
		if idx01.Frame > idx00.Frame {
			return (idx01.Frame - idx00.Frame).ToDuration()
		}
	}

	// Use PREGAP field
	return t.Pregap.ToDuration()
}

// Duration calculates the track duration given the start of the next track
// If this is the last track, nextTrackStart should be the end position
func (t *Track) Duration(nextTrackStart Frame) time.Duration {
	start, err := t.StartPosition()
	if err != nil {
		return 0
	}
	if nextTrackStart <= start {
		return 0
	}
	return (nextTrackStart - start).ToDuration()
}

// HasFlag tests if a specific flag is set
func (t *Track) HasFlag(flag Flags) bool {
	return (t.Flags & flag) != 0
}

// IsCopyPermitted returns true if the DCP (Digital Copy Permitted) flag is set
func (t *Track) IsCopyPermitted() bool {
	return t.HasFlag(Dcp)
}

// IsDataTrack returns true if this is a data track (not AUDIO)
func (t *Track) IsDataTrack() bool {
	return t.TrackDataType != "AUDIO"
}

// HasPreemphasis returns true if the PRE (pre-emphasis) flag is set
func (t *Track) HasPreemphasis() bool {
	return t.HasFlag(Pre)
}

// IsFourChannel returns true if the 4CH (four channel audio) flag is set
func (t *Track) IsFourChannel() bool {
	return t.HasFlag(Four_ch)
}

// HasSCMS returns true if the SCMS (Serial Copy Management System) flag is set
func (t *Track) HasSCMS() bool {
	return t.HasFlag(Scms)
}

// GetBlockSize returns the block size in bytes for this track's data type
func (t *Track) GetBlockSize() int {
	if mode, ok := ValidTrackModes[t.TrackDataType]; ok {
		return mode.BlockSize
	}
	return 0
}

// Frame conversion helpers

// ToDuration converts a Frame to time.Duration
// 75 frames = 1 second (CD standard)
func (f Frame) ToDuration() time.Duration {
	seconds := float64(f) / framesPerSecond
	return time.Duration(seconds * float64(time.Second))
}

// ToSeconds converts a Frame to seconds as a float64
func (f Frame) ToSeconds() float64 {
	return float64(f) / framesPerSecond
}

// DurationToFrame converts a time.Duration to Frame
func DurationToFrame(d time.Duration) Frame {
	seconds := d.Seconds()
	return Frame(seconds * framesPerSecond)
}

// Validation functions

// Validate checks the cuesheet for structural and data validity
func (c *Cuesheet) Validate() []error {
	var errs []error

	// Validate catalog (13 digits)
	if len(c.Catalog) > 0 {
		if err := ValidateCatalog(c.Catalog); err != nil {
			errs = append(errs, err)
		}
	}

	// Validate files
	if len(c.File) == 0 {
		errs = append(errs, strconv.ErrSyntax)
	}

	for _, file := range c.File {
		// Validate file type
		if err := ValidateFileType(file.FileType); err != nil {
			errs = append(errs, err)
		}

		// Validate tracks
		for _, track := range file.Tracks {
			if trackErrs := track.Validate(); len(trackErrs) > 0 {
				errs = append(errs, trackErrs...)
			}
		}
	}

	return errs
}

// Validate checks the track for structural and data validity
func (t *Track) Validate() []error {
	var errs []error

	// Track number range (1-99)
	if t.TrackNumber < 1 || t.TrackNumber > 99 {
		errs = append(errs, strconv.ErrRange)
	}

	// Must have at least INDEX 01
	hasIndex01 := false
	for _, idx := range t.Index {
		if idx.Number == 1 {
			hasIndex01 = true
		}
		// Index range (0-99)
		if idx.Number > 99 {
			errs = append(errs, strconv.ErrRange)
		}
	}
	if !hasIndex01 {
		errs = append(errs, strconv.ErrSyntax)
	}

	// Validate ISRC format
	if len(t.Isrc) > 0 {
		if err := ValidateISRC(t.Isrc); err != nil {
			errs = append(errs, err)
		}
	}

	// Validate track data type
	if err := ValidateTrackDataType(t.TrackDataType); err != nil {
		errs = append(errs, err)
	}

	return errs
}

// ValidateCatalog checks if the catalog number is valid (13 digits)
func ValidateCatalog(catalog string) error {
	if len(catalog) != 13 {
		return strconv.ErrSyntax
	}
	for _, c := range catalog {
		if c < '0' || c > '9' {
			return strconv.ErrSyntax
		}
	}
	return nil
}

// ValidateISRC checks if the ISRC code is valid
// Format: CCOOOOYYSSSSS (12 characters)
//   CC = country code (2 letters)
//   OOOOO = owner code (3 alphanumeric)
//   YY = year (2 digits)
//   SSSSS = serial (5 digits)
func ValidateISRC(isrc string) error {
	if len(isrc) != 12 {
		return strconv.ErrSyntax
	}
	// CC: 2 letters
	if !isLetter(isrc[0]) || !isLetter(isrc[1]) {
		return strconv.ErrSyntax
	}
	// OOOOO: 3 alphanumeric
	for i := 2; i < 5; i++ {
		if !isAlphaNum(isrc[i]) {
			return strconv.ErrSyntax
		}
	}
	// YY: 2 digits
	// SSSSS: 5 digits
	for i := 5; i < 12; i++ {
		if !isDigit(isrc[i]) {
			return strconv.ErrSyntax
		}
	}
	return nil
}

// ValidFileTypes lists valid file types according to CUE specification
var ValidFileTypes = map[string]bool{
	"BINARY":   true,
	"MOTOROLA": true,
	"AIFF":     true,
	"WAVE":     true,
	"MP3":      true,
}

// ValidateFileType checks if the file type is valid
func ValidateFileType(fileType string) error {
	if !ValidFileTypes[fileType] {
		return strconv.ErrSyntax
	}
	return nil
}

// TrackMode describes a track data type with its block size
type TrackMode struct {
	Name      string
	BlockSize int
}

// ValidTrackModes maps track data type names to their specifications
var ValidTrackModes = map[string]TrackMode{
	"AUDIO":        {"AUDIO", 2352},
	"CDG":          {"CDG", 2448},
	"MODE1/2048":   {"MODE1/2048", 2048},
	"MODE1/2352":   {"MODE1/2352", 2352},
	"MODE2/2336":   {"MODE2/2336", 2336},
	"MODE2/2352":   {"MODE2/2352", 2352},
	"CDI/2336":     {"CDI/2336", 2336},
	"CDI/2352":     {"CDI/2352", 2352},
}

// ValidateTrackDataType checks if the track data type is valid
func ValidateTrackDataType(dataType string) error {
	if _, ok := ValidTrackModes[dataType]; !ok {
		return strconv.ErrSyntax
	}
	return nil
}

// Helper functions for validation

func isLetter(c byte) bool {
	return (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z')
}

func isDigit(c byte) bool {
	return c >= '0' && c <= '9'
}

func isAlphaNum(c byte) bool {
	return isLetter(c) || isDigit(c)
}
