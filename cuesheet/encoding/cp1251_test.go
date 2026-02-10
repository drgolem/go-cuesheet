package encoding

import (
	"testing"
)

func TestCP1251ToByte(t *testing.T) {
	tests := []struct {
		name     string
		input    rune
		expected byte
	}{
		// ASCII range
		{"ASCII space", ' ', 0x20},
		{"ASCII A", 'A', 0x41},
		{"ASCII z", 'z', 0x7A},
		{"ASCII 0", '0', 0x30},

		// Cyrillic capital letters
		{"Cyrillic А", 'А', 0xC0}, // U+0410
		{"Cyrillic Б", 'Б', 0xC1},
		{"Cyrillic Я", 'Я', 0xDF}, // U+042F

		// Cyrillic small letters
		{"Cyrillic а", 'а', 0xE0}, // U+0430
		{"Cyrillic б", 'б', 0xE1},
		{"Cyrillic я", 'я', 0xFF}, // U+044F

		// Special Cyrillic characters
		{"Cyrillic Ё", 'Ё', 0xA8},
		{"Cyrillic ё", 'ё', 0xB8},
		{"Cyrillic Ґ", 'Ґ', 0xA5},
		{"Cyrillic ґ", 'ґ', 0xB4},

		// Punctuation
		{"Non-breaking space", '\u00A0', 0xA0},
		{"Copyright", '©', 0xA9},
		{"Left double angle quote", '«', 0xAB},
		{"Right double angle quote", '»', 0xBB},

		// Windows-specific characters
		{"Euro sign", '€', 0x88},
		{"Ellipsis", '…', 0x85},
		{"Em dash", '—', 0x97},
		{"Trademark", '™', 0x99},

		// Characters not in CP1251
		{"Chinese character", '中', 0},
		{"Emoji", '😀', 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CP1251ToByte(tt.input)
			if result != tt.expected {
				t.Errorf("CP1251ToByte(%U) = 0x%02X, want 0x%02X", tt.input, result, tt.expected)
			}
		})
	}
}

func TestDecodeMojibakeFromCP1251(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Russian word Открытие (known working case)",
			input:    "РћС‚РєСЂС‹С‚РёРµ",
			expected: "Открытие",
		},
		{
			name:     "Already correct text",
			input:    "Браво",
			expected: "Браво",
		},
		{
			name:     "ASCII text unchanged",
			input:    "Hello World",
			expected: "Hello World",
		},
		{
			name:     "Empty string",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DecodeMojibakeFromCP1251(tt.input)
			if result != tt.expected {
				t.Errorf("DecodeMojibakeFromCP1251(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestCountCyrillic(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{
			name:     "Pure lowercase Cyrillic",
			input:    "браво",
			expected: 10, // 5 letters * 2 (lowercase weight)
		},
		{
			name:     "Pure uppercase Cyrillic",
			input:    "БРАВО",
			expected: 5, // 5 letters * 1 (uppercase weight)
		},
		{
			name:     "Mixed case Cyrillic",
			input:    "Браво",
			expected: 9, // 1*1 + 4*2
		},
		{
			name:     "No Cyrillic",
			input:    "Hello",
			expected: 0,
		},
		{
			name:     "Empty string",
			input:    "",
			expected: 0,
		},
		{
			name:     "Mixed Cyrillic and ASCII",
			input:    "Hello мир",
			expected: 6, // 3 lowercase * 2
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CountCyrillic(tt.input)
			if result != tt.expected {
				t.Errorf("CountCyrillic(%q) = %d, want %d", tt.input, result, tt.expected)
			}
		})
	}
}

func TestDecodeFromCP1251(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "ASCII unchanged",
			input:    "Hello",
			expected: "Hello",
		},
		{
			name:     "Empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "Basic Latin-1 decode",
			input:    "Test",
			expected: "Test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DecodeFromCP1251(tt.input)
			if result != tt.expected {
				t.Errorf("DecodeFromCP1251(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// Benchmark tests
func BenchmarkCP1251ToByte(b *testing.B) {
	for i := 0; i < b.N; i++ {
		CP1251ToByte('а')
	}
}

func BenchmarkDecodeMojibakeFromCP1251(b *testing.B) {
	input := "Р'СЂР°РІРѕ"
	for i := 0; i < b.N; i++ {
		DecodeMojibakeFromCP1251(input)
	}
}

func BenchmarkCountCyrillic(b *testing.B) {
	input := "Стиляги из Москвы"
	for i := 0; i < b.N; i++ {
		CountCyrillic(input)
	}
}
