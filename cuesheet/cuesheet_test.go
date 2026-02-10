package cuesheet

import (
	"os"
	"reflect"
	"strings"
	"testing"
)

const cueFile = "test.cue"

func TestCuesheet(t *testing.T) {
	actual := Cuesheet{
		Rem: []string{
			"GENRE \"Electronica\"",
			"DATE \"2015\"",
		},
		Catalog:    "1234567890123",
		CdTextFile: "test.cdt",
		Title:      "Test Album Title",
		Performer:  "Test Album Performer",
		SongWriter: "Test Album SongWriter",
		Composer:   "",
		Arranger:   "",
		Message:    "",
		Genre:      "",
		DiscId:     "",
		UpcEan:     "",
		Pregap:     0,
		Postgap:    0,
		File: []File{
			{
				FileName: "test.wav",
				FileType: "WAVE",
				Tracks: []Track{
					{
						TrackNumber:   1,
						TrackDataType: "AUDIO",
						Flags:         None,
						Isrc:          "ABCDE1234567",
						Title:         "Test Title",
						Performer:     "Test Performer",
						SongWriter:    "Test SongWriter",
						Composer:      "",
						Arranger:      "",
						Message:       "",
						Pregap:        0,
						Postgap:       0,
						Index: []TrackIndex{
							{
								Number: 1,
								Frame:  0,
							},
						},
					},
					{
						TrackNumber:   2,
						TrackDataType: "MODE1/2048",
						Flags:         Dcp,
						Isrc:          "ABCDE1234567",
						Title:         "Test Title",
						Performer:     "Test Performer",
						SongWriter:    "Test SongWriter",
						Composer:      "",
						Arranger:      "",
						Message:       "",
						Pregap:        0,
						Postgap:       0,
						Index: []TrackIndex{
							{
								Number: 1,
								Frame:  2715,
							},
						},
					},
				},
			},
		},
	}

	{
		w, err := os.Create(cueFile)
		if err != nil {
			t.Error(err)
		}
		defer w.Close()

		if err := WriteFile(w, &actual); err != nil {
			t.Error(err)
		}
	}

	var expected *Cuesheet

	{
		r, err := os.Open(cueFile)
		if err != nil {
			t.Error(err)
		}
		defer r.Close()
		defer os.Remove(cueFile)

		expected, err = ReadFile(r)
		if err != nil {
			t.Error(err)
		}
	}

	if !reflect.DeepEqual(actual, *expected) {
		t.Errorf("wrong reading data")
	}

	genre := (*expected).Rem[0]
	if ReadString(&genre) != "GENRE" || ReadString(&genre) != "Electronica" {
		t.Errorf("wrong reading data")
	}

	date := (*expected).Rem[1]
	if ReadString(&date) != "DATE" {
		t.Errorf("wrong reading data")
	}
	dateVal, err := ReadUint(&date)
	if err != nil {
		t.Error(err)
	}
	if dateVal != 2015 {
		t.Errorf("wrong reading data")
	}
}

func TestEmptyFile(t *testing.T) {
	reader := strings.NewReader("")
	cuesheet, err := ReadFile(reader)
	if err != nil {
		t.Errorf("expected no error for empty file, got: %v", err)
	}
	if cuesheet == nil {
		t.Error("expected non-nil cuesheet for empty file")
	}
	if len(cuesheet.File) != 0 {
		t.Errorf("expected 0 files, got: %d", len(cuesheet.File))
	}
}

func TestInvalidFrameFormat(t *testing.T) {
	input := `TITLE "Test Album"
PERFORMER "Test Artist"
FILE "test.wav" WAVE
  TRACK 01 AUDIO
    TITLE "Track 1"
    INDEX 01 invalid:frame:format
`
	reader := strings.NewReader(input)
	_, err := ReadFile(reader)
	if err == nil {
		t.Error("expected error for invalid frame format, got nil")
	}
}

func TestInvalidTrackNumber(t *testing.T) {
	input := `FILE "test.wav" WAVE
  TRACK notanumber AUDIO
    INDEX 01 00:00:00
`
	reader := strings.NewReader(input)
	_, err := ReadFile(reader)
	if err == nil {
		t.Error("expected error for invalid track number, got nil")
	}
}

func TestUnquotedStrings(t *testing.T) {
	input := `TITLE TestAlbumNoQuotes
PERFORMER TestArtistNoQuotes
FILE test.wav WAVE
  TRACK 01 AUDIO
    TITLE TrackOneNoQuotes
    PERFORMER TrackArtistNoQuotes
    INDEX 01 00:00:00
`
	reader := strings.NewReader(input)
	cuesheet, err := ReadFile(reader)
	if err != nil {
		t.Errorf("expected no error for unquoted strings, got: %v", err)
	}
	if cuesheet.Title != "TestAlbumNoQuotes" {
		t.Errorf("expected title 'TestAlbumNoQuotes', got: '%s'", cuesheet.Title)
	}
	if cuesheet.Performer != "TestArtistNoQuotes" {
		t.Errorf("expected performer 'TestArtistNoQuotes', got: '%s'", cuesheet.Performer)
	}
	if len(cuesheet.File) != 1 {
		t.Fatalf("expected 1 file, got: %d", len(cuesheet.File))
	}
	if len(cuesheet.File[0].Tracks) != 1 {
		t.Fatalf("expected 1 track, got: %d", len(cuesheet.File[0].Tracks))
	}
	track := cuesheet.File[0].Tracks[0]
	if track.Title != "TrackOneNoQuotes" {
		t.Errorf("expected track title 'TrackOneNoQuotes', got: '%s'", track.Title)
	}
	if track.Performer != "TrackArtistNoQuotes" {
		t.Errorf("expected track performer 'TrackArtistNoQuotes', got: '%s'", track.Performer)
	}
}

func TestQuotedStringsWithSpaces(t *testing.T) {
	input := `TITLE "Test Album With Spaces"
PERFORMER "Test Artist Name"
FILE "test file.wav" WAVE
  TRACK 01 AUDIO
    TITLE "Track One Title"
    PERFORMER "Track Artist Name"
    INDEX 01 00:02:33
`
	reader := strings.NewReader(input)
	cuesheet, err := ReadFile(reader)
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
	if cuesheet.Title != "Test Album With Spaces" {
		t.Errorf("expected title 'Test Album With Spaces', got: '%s'", cuesheet.Title)
	}
	if cuesheet.File[0].FileName != "test file.wav" {
		t.Errorf("expected filename 'test file.wav', got: '%s'", cuesheet.File[0].FileName)
	}
	track := cuesheet.File[0].Tracks[0]
	if track.Title != "Track One Title" {
		t.Errorf("expected track title 'Track One Title', got: '%s'", track.Title)
	}
}

func TestFrameConversion(t *testing.T) {
	tests := []struct {
		input    string
		expected Frame
	}{
		{"00:00:00", 0},
		{"00:00:01", 1},
		{"00:01:00", 75},
		{"01:00:00", 4500},
		{"00:02:15", 165},
	}

	for _, tt := range tests {
		s := tt.input
		frame, err := ReadFrame(&s)
		if err != nil {
			t.Errorf("ReadFrame(%q) error: %v", tt.input, err)
			continue
		}
		if frame != tt.expected {
			t.Errorf("ReadFrame(%q) = %d, expected %d", tt.input, frame, tt.expected)
		}
	}
}

func TestFrameFormatting(t *testing.T) {
	tests := []struct {
		frame    Frame
		expected string
	}{
		{0, "00:00:00"},
		{1, "00:00:01"},
		{75, "00:01:00"},
		{4500, "01:00:00"},
		{165, "00:02:15"},
	}

	for _, tt := range tests {
		result := FormatFrame(tt.frame)
		if result != tt.expected {
			t.Errorf("FormatFrame(%d) = %q, expected %q", tt.frame, result, tt.expected)
		}
	}
}

func TestFlags(t *testing.T) {
	input := `FILE "test.wav" WAVE
  TRACK 01 AUDIO
    FLAGS DCP 4CH PRE SCMS
    INDEX 01 00:00:00
`
	reader := strings.NewReader(input)
	cuesheet, err := ReadFile(reader)
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
	track := cuesheet.File[0].Tracks[0]
	if track.Flags&Dcp == 0 {
		t.Error("expected DCP flag to be set")
	}
	if track.Flags&Four_ch == 0 {
		t.Error("expected Four_ch flag to be set")
	}
	if track.Flags&Pre == 0 {
		t.Error("expected Pre flag to be set")
	}
	if track.Flags&Scms == 0 {
		t.Error("expected Scms flag to be set")
	}
}

func TestMultipleTracks(t *testing.T) {
	input := `TITLE "Multi-Track Album"
FILE "album.wav" WAVE
  TRACK 01 AUDIO
    TITLE "Track 1"
    INDEX 01 00:00:00
  TRACK 02 AUDIO
    TITLE "Track 2"
    INDEX 01 03:25:10
  TRACK 03 AUDIO
    TITLE "Track 3"
    INDEX 01 07:13:45
`
	reader := strings.NewReader(input)
	cuesheet, err := ReadFile(reader)
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
	if len(cuesheet.File) != 1 {
		t.Fatalf("expected 1 file, got: %d", len(cuesheet.File))
	}
	tracks := cuesheet.File[0].Tracks
	if len(tracks) != 3 {
		t.Fatalf("expected 3 tracks, got: %d", len(tracks))
	}
	if tracks[0].Title != "Track 1" {
		t.Errorf("expected track 1 title 'Track 1', got: '%s'", tracks[0].Title)
	}
	if tracks[1].Title != "Track 2" {
		t.Errorf("expected track 2 title 'Track 2', got: '%s'", tracks[1].Title)
	}
	if tracks[2].Title != "Track 3" {
		t.Errorf("expected track 3 title 'Track 3', got: '%s'", tracks[2].Title)
	}
}

func TestRoundTripWithComplexData(t *testing.T) {
	original := Cuesheet{
		Rem:        []string{"GENRE \"Rock\"", "DATE \"2024\"", "COMMENT \"Test\""},
		Catalog:    "1234567890123",
		CdTextFile: "album.cdt",
		Title:      "Complex Album Title",
		Performer:  "Various Artists",
		SongWriter: "Album Composer",
		File: []File{
			{
				FileName: "disc.wav",
				FileType: "WAVE",
				Tracks: []Track{
					{
						TrackNumber:   1,
						TrackDataType: "AUDIO",
						Flags:         Dcp | Pre,
						Isrc:          "USXX12345678",
						Title:         "First Track",
						Performer:     "Artist One",
						SongWriter:    "Writer One",
						Pregap:        150,
						Index: []TrackIndex{
							{Number: 1, Frame: 0},
						},
					},
					{
						TrackNumber:   2,
						TrackDataType: "AUDIO",
						Flags:         None,
						Isrc:          "USXX87654321",
						Title:         "Second Track",
						Performer:     "Artist Two",
						SongWriter:    "Writer Two",
						Index: []TrackIndex{
							{Number: 0, Frame: 15000},
							{Number: 1, Frame: 15150},
						},
					},
				},
			},
		},
	}

	// Write to file
	w, err := os.Create(cueFile)
	if err != nil {
		t.Fatal(err)
	}
	defer w.Close()
	defer os.Remove(cueFile)

	if err := WriteFile(w, &original); err != nil {
		t.Fatal(err)
	}
	w.Close()

	// Read back
	r, err := os.Open(cueFile)
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()

	readBack, err := ReadFile(r)
	if err != nil {
		t.Fatal(err)
	}

	// Compare
	if !reflect.DeepEqual(original, *readBack) {
		t.Error("round-trip data mismatch")
		t.Logf("Original: %+v", original)
		t.Logf("Read back: %+v", *readBack)
	}
}

func TestMultipleFiles(t *testing.T) {
	input := `TITLE "Multi-File Compilation"
FILE "disc1.wav" WAVE
  TRACK 01 AUDIO
    TITLE "Track 1"
    INDEX 01 00:00:00
FILE "disc2.wav" WAVE
  TRACK 02 AUDIO
    TITLE "Track 2"
    INDEX 01 00:00:00
  TRACK 03 AUDIO
    TITLE "Track 3"
    INDEX 01 03:45:22
`
	reader := strings.NewReader(input)
	cuesheet, err := ReadFile(reader)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if len(cuesheet.File) != 2 {
		t.Fatalf("expected 2 files, got: %d", len(cuesheet.File))
	}
	if cuesheet.File[0].FileName != "disc1.wav" {
		t.Errorf("expected first file 'disc1.wav', got: '%s'", cuesheet.File[0].FileName)
	}
	if cuesheet.File[1].FileName != "disc2.wav" {
		t.Errorf("expected second file 'disc2.wav', got: '%s'", cuesheet.File[1].FileName)
	}
	if len(cuesheet.File[0].Tracks) != 1 {
		t.Errorf("expected 1 track in first file, got: %d", len(cuesheet.File[0].Tracks))
	}
	if len(cuesheet.File[1].Tracks) != 2 {
		t.Errorf("expected 2 tracks in second file, got: %d", len(cuesheet.File[1].Tracks))
	}
	if cuesheet.TrackCount() != 3 {
		t.Errorf("expected total 3 tracks, got: %d", cuesheet.TrackCount())
	}
}

func TestValidation(t *testing.T) {
	t.Run("ValidCatalog", func(t *testing.T) {
		if err := ValidateCatalog("1234567890123"); err != nil {
			t.Errorf("expected valid catalog, got error: %v", err)
		}
	})

	t.Run("InvalidCatalogLength", func(t *testing.T) {
		if err := ValidateCatalog("123456789"); err == nil {
			t.Error("expected error for short catalog")
		}
	})

	t.Run("InvalidCatalogChars", func(t *testing.T) {
		if err := ValidateCatalog("12345678901AB"); err == nil {
			t.Error("expected error for non-digit catalog")
		}
	})

	t.Run("ValidISRC", func(t *testing.T) {
		if err := ValidateISRC("USRC17607839"); err != nil {
			t.Errorf("expected valid ISRC, got error: %v", err)
		}
	})

	t.Run("InvalidISRCLength", func(t *testing.T) {
		if err := ValidateISRC("USRC176078"); err == nil {
			t.Error("expected error for short ISRC")
		}
	})

	t.Run("InvalidISRCFormat", func(t *testing.T) {
		if err := ValidateISRC("12RC17607839"); err == nil {
			t.Error("expected error for ISRC with digits in country code")
		}
	})

	t.Run("ValidFileType", func(t *testing.T) {
		if err := ValidateFileType("WAVE"); err != nil {
			t.Errorf("expected valid file type, got error: %v", err)
		}
	})

	t.Run("InvalidFileType", func(t *testing.T) {
		if err := ValidateFileType("FLAC"); err == nil {
			t.Error("expected error for invalid file type")
		}
	})

	t.Run("ValidTrackDataType", func(t *testing.T) {
		if err := ValidateTrackDataType("AUDIO"); err != nil {
			t.Errorf("expected valid track type, got error: %v", err)
		}
		if err := ValidateTrackDataType("MODE1/2048"); err != nil {
			t.Errorf("expected valid track type, got error: %v", err)
		}
	})

	t.Run("InvalidTrackDataType", func(t *testing.T) {
		if err := ValidateTrackDataType("VIDEO"); err == nil {
			t.Error("expected error for invalid track type")
		}
	})
}

func TestHelperMethods(t *testing.T) {
	input := `TITLE "Test Album"
CATALOG 1234567890123
FILE "test.wav" WAVE
  TRACK 01 AUDIO
    TITLE "Track One"
    FLAGS DCP PRE
    ISRC USRC17607839
    INDEX 00 00:00:00
    INDEX 01 00:02:00
  TRACK 02 AUDIO
    TITLE "Track Two"
    INDEX 01 03:30:45
`
	reader := strings.NewReader(input)
	cuesheet, err := ReadFile(reader)
	if err != nil {
		t.Fatalf("ReadFile error: %v", err)
	}

	t.Run("GetTrack", func(t *testing.T) {
		track, err := cuesheet.GetTrack(1)
		if err != nil {
			t.Errorf("GetTrack(1) error: %v", err)
		}
		if track.Title != "Track One" {
			t.Errorf("expected 'Track One', got: '%s'", track.Title)
		}

		_, err = cuesheet.GetTrack(99)
		if err == nil {
			t.Error("expected error for non-existent track")
		}
	})

	t.Run("TrackCount", func(t *testing.T) {
		count := cuesheet.TrackCount()
		if count != 2 {
			t.Errorf("expected 2 tracks, got: %d", count)
		}
	})

	t.Run("GetIndex", func(t *testing.T) {
		track, _ := cuesheet.GetTrack(1)
		idx, err := track.GetIndex(1)
		if err != nil {
			t.Errorf("GetIndex(1) error: %v", err)
		}
		if idx.Frame != 150 { // 00:02:00 = 2*75 = 150 frames
			t.Errorf("expected frame 150, got: %d", idx.Frame)
		}
	})

	t.Run("HasPregap", func(t *testing.T) {
		track, _ := cuesheet.GetTrack(1)
		if !track.HasPregap() {
			t.Error("expected track to have pregap")
		}

		track2, _ := cuesheet.GetTrack(2)
		if track2.HasPregap() {
			t.Error("expected track 2 to not have pregap")
		}
	})

	t.Run("PregapDuration", func(t *testing.T) {
		track, _ := cuesheet.GetTrack(1)
		duration := track.PregapDuration()
		expected := Frame(150).ToDuration() // 2 seconds
		if duration != expected {
			t.Errorf("expected duration %v, got: %v", expected, duration)
		}
	})

	t.Run("HasFlag", func(t *testing.T) {
		track, _ := cuesheet.GetTrack(1)
		if !track.HasFlag(Dcp) {
			t.Error("expected DCP flag")
		}
		if !track.HasFlag(Pre) {
			t.Error("expected PRE flag")
		}
		if track.HasFlag(Four_ch) {
			t.Error("did not expect 4CH flag")
		}
	})

	t.Run("FlagHelpers", func(t *testing.T) {
		track, _ := cuesheet.GetTrack(1)
		if !track.IsCopyPermitted() {
			t.Error("expected copy permitted")
		}
		if !track.HasPreemphasis() {
			t.Error("expected preemphasis")
		}
		if track.IsFourChannel() {
			t.Error("did not expect four channel")
		}
		if track.IsDataTrack() {
			t.Error("expected audio track, not data track")
		}
	})

	t.Run("GetBlockSize", func(t *testing.T) {
		track, _ := cuesheet.GetTrack(1)
		size := track.GetBlockSize()
		if size != 2352 {
			t.Errorf("expected block size 2352 for AUDIO, got: %d", size)
		}
	})

	t.Run("FrameConversion", func(t *testing.T) {
		frame := Frame(150) // 2 seconds
		duration := frame.ToDuration()
		seconds := frame.ToSeconds()

		if seconds != 2.0 {
			t.Errorf("expected 2.0 seconds, got: %f", seconds)
		}

		backToFrame := DurationToFrame(duration)
		if backToFrame != frame {
			t.Errorf("expected frame %d, got: %d", frame, backToFrame)
		}
	})
}

func TestValidateCuesheet(t *testing.T) {
	t.Run("ValidCuesheet", func(t *testing.T) {
		cuesheet := Cuesheet{
			Catalog: "1234567890123",
			File: []File{
				{
					FileName: "test.wav",
					FileType: "WAVE",
					Tracks: []Track{
						{
							TrackNumber:   1,
							TrackDataType: "AUDIO",
							Isrc:          "USRC17607839",
							Index: []TrackIndex{
								{Number: 1, Frame: 0},
							},
						},
					},
				},
			},
		}
		errs := cuesheet.Validate()
		if len(errs) > 0 {
			t.Errorf("expected no errors, got: %v", errs)
		}
	})

	t.Run("InvalidCatalog", func(t *testing.T) {
		cuesheet := Cuesheet{
			Catalog: "123",
			File: []File{
				{
					FileName: "test.wav",
					FileType: "WAVE",
					Tracks: []Track{
						{
							TrackNumber:   1,
							TrackDataType: "AUDIO",
							Index: []TrackIndex{
								{Number: 1, Frame: 0},
							},
						},
					},
				},
			},
		}
		errs := cuesheet.Validate()
		if len(errs) == 0 {
			t.Error("expected validation errors")
		}
	})

	t.Run("InvalidTrackNumber", func(t *testing.T) {
		cuesheet := Cuesheet{
			File: []File{
				{
					FileName: "test.wav",
					FileType: "WAVE",
					Tracks: []Track{
						{
							TrackNumber:   100,
							TrackDataType: "AUDIO",
							Index: []TrackIndex{
								{Number: 1, Frame: 0},
							},
						},
					},
				},
			},
		}
		errs := cuesheet.Validate()
		if len(errs) == 0 {
			t.Error("expected validation errors for invalid track number")
		}
	})

	t.Run("MissingIndex01", func(t *testing.T) {
		cuesheet := Cuesheet{
			File: []File{
				{
					FileName: "test.wav",
					FileType: "WAVE",
					Tracks: []Track{
						{
							TrackNumber:   1,
							TrackDataType: "AUDIO",
							Index: []TrackIndex{
								{Number: 0, Frame: 0},
							},
						},
					},
				},
			},
		}
		errs := cuesheet.Validate()
		if len(errs) == 0 {
			t.Error("expected validation errors for missing INDEX 01")
		}
	})
}

func TestRemParsing(t *testing.T) {
	t.Run("ParseRemDate", func(t *testing.T) {
		field, ok := ParseRemComment("DATE \"2024\"")
		if !ok {
			t.Fatal("expected successful parse")
		}
		if field.Type != RemDate {
			t.Errorf("expected RemDate type, got: %v", field.Type)
		}
		if field.Key != "DATE" {
			t.Errorf("expected key 'DATE', got: '%s'", field.Key)
		}
		if field.Value != "2024" {
			t.Errorf("expected value '2024', got: '%s'", field.Value)
		}
	})

	t.Run("ParseRemGenre", func(t *testing.T) {
		field, ok := ParseRemComment("GENRE \"Rock\"")
		if !ok {
			t.Fatal("expected successful parse")
		}
		if field.Type != RemGenre {
			t.Errorf("expected RemGenre type, got: %v", field.Type)
		}
		if field.Value != "Rock" {
			t.Errorf("expected value 'Rock', got: '%s'", field.Value)
		}
	})

	t.Run("ParseRemDiscNumber", func(t *testing.T) {
		field, ok := ParseRemComment("DISCNUMBER 1")
		if !ok {
			t.Fatal("expected successful parse")
		}
		if field.Type != RemDiscNumber {
			t.Errorf("expected RemDiscNumber type, got: %v", field.Type)
		}
		if field.Value != "1" {
			t.Errorf("expected value '1', got: '%s'", field.Value)
		}
	})

	t.Run("ParseRemReplayGain", func(t *testing.T) {
		field, ok := ParseRemComment("REPLAYGAIN_ALBUM_GAIN -6.2 dB")
		if !ok {
			t.Fatal("expected successful parse")
		}
		if field.Type != RemReplayGainAlbumGain {
			t.Errorf("expected RemReplayGainAlbumGain type, got: %v", field.Type)
		}
		if field.Value != "-6.2 dB" {
			t.Errorf("expected value '-6.2 dB', got: '%s'", field.Value)
		}
	})

	t.Run("GetRemValue", func(t *testing.T) {
		cuesheet := Cuesheet{
			Rem: []string{
				"GENRE \"Electronica\"",
				"DATE \"2015\"",
				"COMMENT \"Test album\"",
			},
		}

		genre, ok := cuesheet.GetRemValue(RemGenre)
		if !ok {
			t.Error("expected to find genre")
		}
		if genre != "Electronica" {
			t.Errorf("expected 'Electronica', got: '%s'", genre)
		}

		date, ok := cuesheet.GetRemValue(RemDate)
		if !ok {
			t.Error("expected to find date")
		}
		if date != "2015" {
			t.Errorf("expected '2015', got: '%s'", date)
		}
	})

	t.Run("GetRemByKey", func(t *testing.T) {
		cuesheet := Cuesheet{
			Rem: []string{
				"GENRE \"Rock\"",
				"CUSTOM_FIELD Value",
			},
		}

		value, ok := cuesheet.GetRemByKey("CUSTOM_FIELD")
		if !ok {
			t.Error("expected to find custom field")
		}
		if value != "Value" {
			t.Errorf("expected 'Value', got: '%s'", value)
		}
	})

	t.Run("GetRemFields", func(t *testing.T) {
		cuesheet := Cuesheet{
			Rem: []string{
				"DATE \"2024\"",
				"GENRE \"Jazz\"",
				"DISCNUMBER 2",
			},
		}

		fields := cuesheet.GetRemFields()
		if len(fields) != 3 {
			t.Errorf("expected 3 fields, got: %d", len(fields))
		}
		if fields[0].Type != RemDate {
			t.Error("expected first field to be RemDate")
		}
		if fields[1].Type != RemGenre {
			t.Error("expected second field to be RemGenre")
		}
		if fields[2].Type != RemDiscNumber {
			t.Error("expected third field to be RemDiscNumber")
		}
	})
}

func TestCDTextFields(t *testing.T) {
	input := `TITLE "Album Title"
PERFORMER "Album Artist"
SONGWRITER "Album Writer"
COMPOSER "Album Composer"
ARRANGER "Album Arranger"
MESSAGE "Album Message"
GENRE "Rock"
DISC_ID "ABC123"
UPC_EAN "1234567890123"
FILE "test.wav" WAVE
  TRACK 01 AUDIO
    TITLE "Track Title"
    PERFORMER "Track Artist"
    SONGWRITER "Track Writer"
    COMPOSER "Track Composer"
    ARRANGER "Track Arranger"
    MESSAGE "Track Message"
    INDEX 01 00:00:00
`
	reader := strings.NewReader(input)
	cuesheet, err := ReadFile(reader)
	if err != nil {
		t.Fatalf("ReadFile error: %v", err)
	}

	// Test album-level CD-TEXT
	if cuesheet.Title != "Album Title" {
		t.Errorf("expected 'Album Title', got: '%s'", cuesheet.Title)
	}
	if cuesheet.Composer != "Album Composer" {
		t.Errorf("expected 'Album Composer', got: '%s'", cuesheet.Composer)
	}
	if cuesheet.Arranger != "Album Arranger" {
		t.Errorf("expected 'Album Arranger', got: '%s'", cuesheet.Arranger)
	}
	if cuesheet.Message != "Album Message" {
		t.Errorf("expected 'Album Message', got: '%s'", cuesheet.Message)
	}
	if cuesheet.Genre != "Rock" {
		t.Errorf("expected 'Rock', got: '%s'", cuesheet.Genre)
	}
	if cuesheet.DiscId != "ABC123" {
		t.Errorf("expected 'ABC123', got: '%s'", cuesheet.DiscId)
	}
	if cuesheet.UpcEan != "1234567890123" {
		t.Errorf("expected '1234567890123', got: '%s'", cuesheet.UpcEan)
	}

	// Test track-level CD-TEXT
	track := cuesheet.File[0].Tracks[0]
	if track.Composer != "Track Composer" {
		t.Errorf("expected 'Track Composer', got: '%s'", track.Composer)
	}
	if track.Arranger != "Track Arranger" {
		t.Errorf("expected 'Track Arranger', got: '%s'", track.Arranger)
	}
	if track.Message != "Track Message" {
		t.Errorf("expected 'Track Message', got: '%s'", track.Message)
	}

	// Test round-trip with new fields
	t.Run("RoundTripCDText", func(t *testing.T) {
		original := Cuesheet{
			Title:      "Test Album",
			Performer:  "Test Artist",
			Composer:   "Test Composer",
			Arranger:   "Test Arranger",
			Genre:      "Jazz",
			Message:    "Test Message",
			DiscId:     "DISC001",
			UpcEan:     "9876543210987",
			File: []File{
				{
					FileName: "audio.wav",
					FileType: "WAVE",
					Tracks: []Track{
						{
							TrackNumber:   1,
							TrackDataType: "AUDIO",
							Title:         "Track 1",
							Composer:      "Track Composer",
							Arranger:      "Track Arranger",
							Message:       "Track Message",
							Index: []TrackIndex{
								{Number: 1, Frame: 0},
							},
						},
					},
				},
			},
		}

		w, err := os.Create(cueFile)
		if err != nil {
			t.Fatal(err)
		}
		defer w.Close()
		defer os.Remove(cueFile)

		if err := WriteFile(w, &original); err != nil {
			t.Fatal(err)
		}
		w.Close()

		r, err := os.Open(cueFile)
		if err != nil {
			t.Fatal(err)
		}
		defer r.Close()

		readBack, err := ReadFile(r)
		if err != nil {
			t.Fatal(err)
		}

		if !reflect.DeepEqual(original, *readBack) {
			t.Error("round-trip data mismatch for CD-TEXT fields")
			t.Logf("Original: %+v", original)
			t.Logf("Read back: %+v", *readBack)
		}
	})
}

func TestSample1Cue(t *testing.T) {
	// Read the sample CUE file from testdata
	file, err := os.Open("testdata/sample_1.cue")
	if err != nil {
		t.Fatalf("failed to open sample_1.cue: %v", err)
	}
	defer file.Close()

	cuesheet, err := ReadFile(file)
	if err != nil {
		t.Fatalf("failed to parse sample_1.cue: %v", err)
	}

	// Test album-level metadata
	t.Run("AlbumMetadata", func(t *testing.T) {
		if cuesheet.Title != "Album Title" {
			t.Errorf("expected album title 'Album Title', got: '%s'", cuesheet.Title)
		}
		if cuesheet.Performer != "Artist Name" {
			t.Errorf("expected album performer 'Artist Name', got: '%s'", cuesheet.Performer)
		}
	})

	// Test REM comments
	t.Run("RemComments", func(t *testing.T) {
		if len(cuesheet.Rem) != 2 {
			t.Fatalf("expected 2 REM comments, got: %d", len(cuesheet.Rem))
		}

		// Test REM parsing
		genre, ok := cuesheet.GetRemValue(RemGenre)
		if !ok {
			t.Error("expected to find GENRE in REM")
		}
		if genre != "Ambient" {
			t.Errorf("expected genre 'Ambient', got: '%s'", genre)
		}

		date, ok := cuesheet.GetRemValue(RemDate)
		if !ok {
			t.Error("expected to find DATE in REM")
		}
		if date != "2025" {
			t.Errorf("expected date '2025', got: '%s'", date)
		}
	})

	// Test file information
	t.Run("FileInfo", func(t *testing.T) {
		if len(cuesheet.File) != 1 {
			t.Fatalf("expected 1 file, got: %d", len(cuesheet.File))
		}

		file := cuesheet.File[0]
		if file.FileName != "Full_Mix.wav" {
			t.Errorf("expected filename 'Full_Mix.wav', got: '%s'", file.FileName)
		}
		if file.FileType != "WAVE" {
			t.Errorf("expected file type 'WAVE', got: '%s'", file.FileType)
		}
	})

	// Test tracks
	t.Run("Tracks", func(t *testing.T) {
		if cuesheet.TrackCount() != 3 {
			t.Fatalf("expected 3 tracks, got: %d", cuesheet.TrackCount())
		}

		tracks := cuesheet.File[0].Tracks

		// Track 1
		if tracks[0].TrackNumber != 1 {
			t.Errorf("expected track number 1, got: %d", tracks[0].TrackNumber)
		}
		if tracks[0].Title != "First Song" {
			t.Errorf("expected track 1 title 'First Song', got: '%s'", tracks[0].Title)
		}
		if tracks[0].Performer != "Artist Name" {
			t.Errorf("expected track 1 performer 'Artist Name', got: '%s'", tracks[0].Performer)
		}
		if tracks[0].TrackDataType != "AUDIO" {
			t.Errorf("expected track 1 type 'AUDIO', got: '%s'", tracks[0].TrackDataType)
		}

		// Track 2
		if tracks[1].TrackNumber != 2 {
			t.Errorf("expected track number 2, got: %d", tracks[1].TrackNumber)
		}
		if tracks[1].Title != "Second Song" {
			t.Errorf("expected track 2 title 'Second Song', got: '%s'", tracks[1].Title)
		}

		// Track 3
		if tracks[2].TrackNumber != 3 {
			t.Errorf("expected track number 3, got: %d", tracks[2].TrackNumber)
		}
		if tracks[2].Title != "Third Song" {
			t.Errorf("expected track 3 title 'Third Song', got: '%s'", tracks[2].Title)
		}
	})

	// Test index positions and frame calculations
	t.Run("IndexPositions", func(t *testing.T) {
		tracks := cuesheet.File[0].Tracks

		// Track 1: INDEX 01 00:00:00 = 0 frames
		idx1, err := tracks[0].GetIndex(1)
		if err != nil {
			t.Errorf("track 1 missing INDEX 01: %v", err)
		} else if idx1.Frame != 0 {
			t.Errorf("track 1 INDEX 01: expected 0 frames, got: %d", idx1.Frame)
		}

		// Track 2: INDEX 01 05:30:00 = (5*60 + 30) * 75 = 24750 frames
		idx2, err := tracks[1].GetIndex(1)
		if err != nil {
			t.Errorf("track 2 missing INDEX 01: %v", err)
		} else {
			expectedFrames := Frame((5*60 + 30) * 75) // 5 min 30 sec
			if idx2.Frame != expectedFrames {
				t.Errorf("track 2 INDEX 01: expected %d frames, got: %d", expectedFrames, idx2.Frame)
			}
		}

		// Track 3: INDEX 01 10:15:50 = (10*60 + 15) * 75 + 50 = 46175 frames
		idx3, err := tracks[2].GetIndex(1)
		if err != nil {
			t.Errorf("track 3 missing INDEX 01: %v", err)
		} else {
			expectedFrames := Frame((10*60+15)*75 + 50) // 10 min 15 sec 50 frames
			if idx3.Frame != expectedFrames {
				t.Errorf("track 3 INDEX 01: expected %d frames, got: %d", expectedFrames, idx3.Frame)
			}
		}
	})

	// Test track durations
	t.Run("TrackDurations", func(t *testing.T) {
		tracks := cuesheet.File[0].Tracks

		// Track 1 duration: from 00:00:00 to 05:30:00 = 5.5 minutes
		track1Duration := tracks[0].Duration(tracks[1].Index[0].Frame)
		expectedDuration1 := Frame((5*60 + 30) * 75).ToDuration()
		if track1Duration != expectedDuration1 {
			t.Errorf("track 1 duration: expected %v, got: %v", expectedDuration1, track1Duration)
		}

		// Track 2 duration: from 05:30:00 to 10:15:50 = 4 min 45.67 sec
		track2Duration := tracks[1].Duration(tracks[2].Index[0].Frame)
		// (10*60+15)*75+50 - (5*60+30)*75 = 46175 - 24750 = 21425 frames
		expectedDuration2 := Frame(21425).ToDuration()
		if track2Duration != expectedDuration2 {
			t.Errorf("track 2 duration: expected %v, got: %v", expectedDuration2, track2Duration)
		}
	})

	// Test helper methods
	t.Run("HelperMethods", func(t *testing.T) {
		// GetTrack by number
		track2, err := cuesheet.GetTrack(2)
		if err != nil {
			t.Errorf("GetTrack(2) failed: %v", err)
		} else if track2.Title != "Second Song" {
			t.Errorf("GetTrack(2) returned wrong track: '%s'", track2.Title)
		}

		// Test frame to time conversions
		track1 := cuesheet.File[0].Tracks[0]
		startPos, err := track1.StartPosition()
		if err != nil {
			t.Errorf("StartPosition() failed: %v", err)
		}
		if startPos != 0 {
			t.Errorf("track 1 start position: expected 0, got: %d", startPos)
		}

		// Test that track 1 has no pregap
		if track1.HasPregap() {
			t.Error("track 1 should not have pregap (INDEX 00)")
		}
	})

	// Test validation
	t.Run("Validation", func(t *testing.T) {
		errs := cuesheet.Validate()
		if len(errs) > 0 {
			t.Errorf("validation failed: %v", errs)
		}
	})

	// Test round-trip (read -> write -> read)
	t.Run("RoundTrip", func(t *testing.T) {
		// Write to temp file
		tmpFile, err := os.CreateTemp("", "sample_1_roundtrip_*.cue")
		if err != nil {
			t.Fatalf("failed to create temp file: %v", err)
		}
		defer os.Remove(tmpFile.Name())
		defer tmpFile.Close()

		if err := WriteFile(tmpFile, cuesheet); err != nil {
			t.Fatalf("WriteFile failed: %v", err)
		}
		tmpFile.Close()

		// Read back
		readFile, err := os.Open(tmpFile.Name())
		if err != nil {
			t.Fatalf("failed to open temp file: %v", err)
		}
		defer readFile.Close()

		cuesheetReadBack, err := ReadFile(readFile)
		if err != nil {
			t.Fatalf("ReadFile failed on round-trip: %v", err)
		}

		// Compare key fields
		if cuesheetReadBack.Title != cuesheet.Title {
			t.Error("round-trip: album title mismatch")
		}
		if cuesheetReadBack.TrackCount() != cuesheet.TrackCount() {
			t.Error("round-trip: track count mismatch")
		}
		if len(cuesheetReadBack.Rem) != len(cuesheet.Rem) {
			t.Error("round-trip: REM count mismatch")
		}

		// Deep equality check
		if !reflect.DeepEqual(cuesheet, cuesheetReadBack) {
			t.Error("round-trip: complete data mismatch")
			t.Logf("Original: %+v", cuesheet)
			t.Logf("Read back: %+v", cuesheetReadBack)
		}
	})
}

func TestSample2Cue(t *testing.T) {
	file, err := os.Open("testdata/sample_2.cue")
	if err != nil {
		t.Fatalf("failed to open sample_2.cue: %v", err)
	}
	defer file.Close()

	cuesheet, err := ReadFile(file)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}

	t.Run("AlbumMetadata", func(t *testing.T) {
		if cuesheet.Title != "Cold Spring Harbor" {
			t.Errorf("expected title 'Cold Spring Harbor', got '%s'", cuesheet.Title)
		}
		if cuesheet.Performer != "Billy Joel" {
			t.Errorf("expected performer 'Billy Joel', got '%s'", cuesheet.Performer)
		}
	})

	t.Run("AlbumRemComments", func(t *testing.T) {
		// Check that we have album-level REM comments
		// GENRE, DATE, DISCID, REPLAYGAIN_ALBUM_GAIN, REPLAYGAIN_ALBUM_PEAK
		if len(cuesheet.Rem) != 5 {
			t.Errorf("expected 5 album-level REM comments, got %d", len(cuesheet.Rem))
		}

		// Test structured REM parsing
		genre, found := cuesheet.GetRemValue(RemGenre)
		if !found {
			t.Error("expected to find GENRE REM")
		}
		if genre != "Rock" {
			t.Errorf("expected GENRE 'Rock', got '%s'", genre)
		}

		date, found := cuesheet.GetRemValue(RemDate)
		if !found {
			t.Error("expected to find DATE REM")
		}
		if date != "2011" {
			t.Errorf("expected DATE '2011', got '%s'", date)
		}

		// Check DISCID
		expectedDiscId := "7E07210A"
		foundDiscId := false
		for _, rem := range cuesheet.Rem {
			if strings.Contains(rem, "DISCID "+expectedDiscId) {
				foundDiscId = true
				break
			}
		}
		if !foundDiscId {
			t.Errorf("expected to find DISCID %s in REM comments", expectedDiscId)
		}

		// Check album ReplayGain
		albumGain, found := cuesheet.GetRemValue(RemReplayGainAlbumGain)
		if !found {
			t.Error("expected to find REPLAYGAIN_ALBUM_GAIN REM")
		}
		if albumGain != "-7.11 dB" {
			t.Errorf("expected REPLAYGAIN_ALBUM_GAIN '-7.11 dB', got '%s'", albumGain)
		}

		albumPeak, found := cuesheet.GetRemValue(RemReplayGainAlbumPeak)
		if !found {
			t.Error("expected to find REPLAYGAIN_ALBUM_PEAK REM")
		}
		if albumPeak != "0.995819" {
			t.Errorf("expected REPLAYGAIN_ALBUM_PEAK '0.995819', got '%s'", albumPeak)
		}
	})

	t.Run("MultipleFiles", func(t *testing.T) {
		// This CUE file has one FILE per track (10 files total)
		if len(cuesheet.File) != 10 {
			t.Errorf("expected 10 FILE entries, got %d", len(cuesheet.File))
		}

		// Check first file
		if cuesheet.File[0].FileName != "01 - Billy Joel - She's Got A Way.flac" {
			t.Errorf("unexpected first file name: %s", cuesheet.File[0].FileName)
		}
		if cuesheet.File[0].FileType != "WAVE" {
			t.Errorf("expected file type WAVE, got %s", cuesheet.File[0].FileType)
		}

		// Check last file
		if cuesheet.File[9].FileName != "10 - Billy Joel - Got To Begin Again.flac" {
			t.Errorf("unexpected last file name: %s", cuesheet.File[9].FileName)
		}

		// Each file should have exactly one track
		for i, file := range cuesheet.File {
			if len(file.Tracks) != 1 {
				t.Errorf("file %d should have 1 track, got %d", i, len(file.Tracks))
			}
		}
	})

	t.Run("TrackCount", func(t *testing.T) {
		if cuesheet.TrackCount() != 10 {
			t.Errorf("expected 10 tracks, got %d", cuesheet.TrackCount())
		}
	})

	t.Run("TrackMetadata", func(t *testing.T) {
		// Test track 1
		track1, err := cuesheet.GetTrack(1)
		if err != nil {
			t.Fatalf("failed to get track 1: %v", err)
		}
		if track1.Title != "She's Got A Way" {
			t.Errorf("track 1 title: expected 'She's Got A Way', got '%s'", track1.Title)
		}
		if track1.Isrc != "USSM11100711" {
			t.Errorf("track 1 ISRC: expected 'USSM11100711', got '%s'", track1.Isrc)
		}

		// Test track 5
		track5, err := cuesheet.GetTrack(5)
		if err != nil {
			t.Fatalf("failed to get track 5: %v", err)
		}
		if track5.Title != "Falling of the Rain" {
			t.Errorf("track 5 title: expected 'Falling of the Rain', got '%s'", track5.Title)
		}
		if track5.Isrc != "USSM11100715" {
			t.Errorf("track 5 ISRC: expected 'USSM11100715', got '%s'", track5.Isrc)
		}

		// Test track 10
		track10, err := cuesheet.GetTrack(10)
		if err != nil {
			t.Fatalf("failed to get track 10: %v", err)
		}
		if track10.Title != "Got To Begin Again" {
			t.Errorf("track 10 title: expected 'Got To Begin Again', got '%s'", track10.Title)
		}
		if track10.Isrc != "USSM11100720" {
			t.Errorf("track 10 ISRC: expected 'USSM11100720', got '%s'", track10.Isrc)
		}
	})

	// Note: Track-level REM comments are currently ignored by the parser
	// (see readTrack function case "REM": // ignore comment inside of track)
	// So we cannot test track-level ReplayGain values

	t.Run("IndexPositions", func(t *testing.T) {
		// All tracks have INDEX 01 at 00:00:00 since each has its own file
		for i := 1; i <= 10; i++ {
			track, _ := cuesheet.GetTrack(uint(i))
			if len(track.Index) != 1 {
				t.Errorf("track %d: expected 1 index, got %d", i, len(track.Index))
				continue
			}

			index01, err := track.GetIndex(1)
			if err != nil {
				t.Errorf("track %d: INDEX 01 not found: %v", i, err)
				continue
			}
			if index01.Frame != 0 {
				t.Errorf("track %d: INDEX 01 should be 00:00:00, got %d frames", i, index01.Frame)
			}
		}
	})

	t.Run("HelperMethods", func(t *testing.T) {
		// Test TrackCount
		if cuesheet.TrackCount() != 10 {
			t.Errorf("TrackCount: expected 10, got %d", cuesheet.TrackCount())
		}

		// Test GetTrack
		track1, err := cuesheet.GetTrack(1)
		if err != nil {
			t.Errorf("GetTrack(1) failed: %v", err)
		}
		if track1.TrackNumber != 1 {
			t.Errorf("GetTrack(1): expected track number 1, got %d", track1.TrackNumber)
		}

		// Test GetTrack with invalid number
		_, err = cuesheet.GetTrack(99)
		if err == nil {
			t.Error("GetTrack(99) should return error")
		}

		// Test HasPregap (should all be false for this CUE)
		for i := 1; i <= 10; i++ {
			track, _ := cuesheet.GetTrack(uint(i))
			if track.HasPregap() {
				t.Errorf("track %d: HasPregap should be false", i)
			}
		}

		// Test GetIndex
		track5, _ := cuesheet.GetTrack(5)
		index01, err := track5.GetIndex(1)
		if err != nil {
			t.Errorf("track 5: INDEX 01 not found: %v", err)
		} else if index01.Frame != 0 {
			t.Errorf("track 5: INDEX 01 should be 0, got %d", index01.Frame)
		}

		// Test GetIndex with non-existent index
		_, err = track5.GetIndex(0)
		if err == nil {
			t.Error("track 5: INDEX 00 should not exist")
		}
	})

	t.Run("Validation", func(t *testing.T) {
		// Validate the entire cuesheet
		errs := cuesheet.Validate()
		if len(errs) > 0 {
			t.Errorf("validation failed with %d errors:", len(errs))
			for _, err := range errs {
				t.Logf("  - %v", err)
			}
		}

		// Test ISRC validation for all tracks
		for i := 1; i <= 10; i++ {
			track, _ := cuesheet.GetTrack(uint(i))
			if err := ValidateISRC(track.Isrc); err != nil {
				t.Errorf("track %d ISRC validation failed: %v", i, err)
			}
		}
	})

	t.Run("RoundTrip", func(t *testing.T) {
		// Write to temp file
		tmpFile, err := os.CreateTemp("", "sample_2_roundtrip_*.cue")
		if err != nil {
			t.Fatalf("failed to create temp file: %v", err)
		}
		defer os.Remove(tmpFile.Name())

		if err := WriteFile(tmpFile, cuesheet); err != nil {
			t.Fatalf("WriteFile failed: %v", err)
		}
		tmpFile.Close()

		// Read back
		readFile, err := os.Open(tmpFile.Name())
		if err != nil {
			t.Fatalf("failed to open temp file: %v", err)
		}
		defer readFile.Close()

		cuesheetReadBack, err := ReadFile(readFile)
		if err != nil {
			t.Fatalf("ReadFile failed on round-trip: %v", err)
		}

		// Compare key fields
		if cuesheetReadBack.Title != cuesheet.Title {
			t.Error("round-trip: album title mismatch")
		}
		if cuesheetReadBack.Performer != cuesheet.Performer {
			t.Error("round-trip: album performer mismatch")
		}
		if cuesheetReadBack.TrackCount() != cuesheet.TrackCount() {
			t.Error("round-trip: track count mismatch")
		}
		if len(cuesheetReadBack.File) != len(cuesheet.File) {
			t.Errorf("round-trip: file count mismatch (expected %d, got %d)",
				len(cuesheet.File), len(cuesheetReadBack.File))
		}

		// Check all tracks have correct ISRC codes
		for i := 1; i <= 10; i++ {
			origTrack, _ := cuesheet.GetTrack(uint(i))
			readTrack, _ := cuesheetReadBack.GetTrack(uint(i))
			if readTrack.Isrc != origTrack.Isrc {
				t.Errorf("round-trip: track %d ISRC mismatch (expected %s, got %s)",
					i, origTrack.Isrc, readTrack.Isrc)
			}
		}

		// Deep equality check
		if !reflect.DeepEqual(cuesheet, cuesheetReadBack) {
			t.Error("round-trip: complete data mismatch")
			t.Logf("Original REM count: %d", len(cuesheet.Rem))
			t.Logf("Read back REM count: %d", len(cuesheetReadBack.Rem))
			t.Logf("Original File count: %d", len(cuesheet.File))
			t.Logf("Read back File count: %d", len(cuesheetReadBack.File))
		}
	})
}
