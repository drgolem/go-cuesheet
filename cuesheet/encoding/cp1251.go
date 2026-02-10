// Package encoding provides character encoding utilities for CUE sheet processing.
package encoding

// CP1251ToByte converts a Unicode character to its CP1251 (Windows-1251) byte value.
// Returns 0 if the character is not in CP1251 encoding (except for actual 0x00).
//
// CP1251 is a Windows code page for Cyrillic scripts, commonly used in CUE files
// from Eastern European and Russian sources.
func CP1251ToByte(r rune) byte {
	// 0x00-0x7F: ASCII range (same in CP1251 and Unicode)
	if r <= 0x7F {
		return byte(r)
	}

	// 0x80-0xFF: Cyrillic range in CP1251
	// Map Unicode Cyrillic characters to CP1251 byte values
	switch {
	// Cyrillic capital letters А-Я (U+0410 - U+042F) → 0xC0-0xDF
	case r >= 0x0410 && r <= 0x042F:
		return byte(0xC0 + (r - 0x0410))

	// Cyrillic small letters а-я (U+0430 - U+044F) → 0xE0-0xFF
	case r >= 0x0430 && r <= 0x044F:
		return byte(0xE0 + (r - 0x0430))

	// Special Cyrillic characters
	case r == 0x0401: // Ё
		return 0xA8
	case r == 0x0451: // ё
		return 0xB8
	case r == 0x0490: // Ґ
		return 0xA5
	case r == 0x0491: // ґ
		return 0xB4
	case r == 0x0404: // Є
		return 0xAA
	case r == 0x0454: // є
		return 0xBA
	case r == 0x0406: // І
		return 0xB2
	case r == 0x0456: // і
		return 0xB3
	case r == 0x0407: // Ї
		return 0xAF
	case r == 0x0457: // ї
		return 0xBF

	// Common punctuation and symbols in CP1251
	case r == 0x00A0: // non-breaking space
		return 0xA0
	case r == 0x00A4: // currency sign
		return 0xA4
	case r == 0x00A6: // broken bar
		return 0xA6
	case r == 0x00A7: // section sign
		return 0xA7
	case r == 0x00A9: // copyright
		return 0xA9
	case r == 0x00AB: // left double angle quote
		return 0xAB
	case r == 0x00AC: // not sign
		return 0xAC
	case r == 0x00AD: // soft hyphen
		return 0xAD
	case r == 0x00AE: // registered
		return 0xAE
	case r == 0x00B0: // degree
		return 0xB0
	case r == 0x00B1: // plus-minus
		return 0xB1
	case r == 0x00B5: // micro
		return 0xB5
	case r == 0x00B6: // paragraph
		return 0xB6
	case r == 0x00B7: // middle dot
		return 0xB7
	case r == 0x00BB: // right double angle quote
		return 0xBB

	// Windows-specific characters in 0x80-0x9F range
	case r == 0x0402: // Ђ
		return 0x80
	case r == 0x0403: // Ѓ
		return 0x81
	case r == 0x201A: // single low quote
		return 0x82
	case r == 0x0453: // ѓ
		return 0x83
	case r == 0x201E: // double low quote
		return 0x84
	case r == 0x2026: // ellipsis
		return 0x85
	case r == 0x2020: // dagger
		return 0x86
	case r == 0x2021: // double dagger
		return 0x87
	case r == 0x20AC: // euro
		return 0x88
	case r == 0x2030: // per mille
		return 0x89
	case r == 0x0409: // Љ
		return 0x8A
	case r == 0x2039: // left single angle quote
		return 0x8B
	case r == 0x040A: // Њ
		return 0x8C
	case r == 0x040C: // Ќ
		return 0x8D
	case r == 0x040B: // Ћ
		return 0x8E
	case r == 0x040F: // Џ
		return 0x8F
	case r == 0x0452: // ђ
		return 0x90
	case r == 0x2018: // left single quote
		return 0x91
	case r == 0x2019: // right single quote
		return 0x92
	case r == 0x201C: // left double quote
		return 0x93
	case r == 0x201D: // right double quote
		return 0x94
	case r == 0x2022: // bullet
		return 0x95
	case r == 0x2013: // en dash
		return 0x96
	case r == 0x2014: // em dash
		return 0x97
	case r == 0x2122: // trademark
		return 0x99
	case r == 0x0459: // љ
		return 0x9A
	case r == 0x203A: // right single angle quote
		return 0x9B
	case r == 0x045A: // њ
		return 0x9C
	case r == 0x045C: // ќ
		return 0x9D
	case r == 0x045B: // ћ
		return 0x9E
	case r == 0x045F: // џ
		return 0x9F
	}

	return 0
}

// DecodeMojibakeFromCP1251 fixes UTF-8 text that was incorrectly read as CP1251 (Windows-1251).
// This is a common issue when UTF-8 encoded files are opened with CP1251 encoding,
// resulting in "mojibake" (garbled text).
//
// The function attempts to reverse the misencoding by:
// 1. Converting each Unicode character back to its CP1251 byte value
// 2. Reinterpreting those bytes as UTF-8
// 3. Validating the result by counting Cyrillic characters
//
// If the decoded text has more Cyrillic characters than the original,
// it's likely the correct decoding. Otherwise, returns the original string.
func DecodeMojibakeFromCP1251(mojibake string) string {
	var utf8Bytes []byte
	allMapped := true

	for _, r := range mojibake {
		// Get the CP1251 byte value for this character
		byteVal := CP1251ToByte(r)
		if byteVal != 0 || r == 0 {
			utf8Bytes = append(utf8Bytes, byteVal)
		} else {
			// Character not in CP1251 - this might not be mojibake
			allMapped = false
			break
		}
	}

	// If we couldn't map all characters, return original
	if !allMapped {
		return mojibake
	}

	decoded := string(utf8Bytes)

	// Check if decoded string looks more like Cyrillic than the original
	// Count Cyrillic characters in both strings
	originalCyrillic := CountCyrillic(mojibake)
	decodedCyrillic := CountCyrillic(decoded)

	// If decoded has more proper Cyrillic small letters (common in text), it's likely correct
	if decodedCyrillic > originalCyrillic {
		return decoded
	}

	return mojibake
}

// CountCyrillic counts the number of Cyrillic characters in a string.
// Lowercase letters are weighted more heavily (2x) as they're more common in normal text.
// This helps determine if a decoded string is more "Cyrillic-like" than the original.
func CountCyrillic(s string) int {
	count := 0
	for _, r := range s {
		// Cyrillic lowercase letters are more common in normal text
		if r >= 0x0430 && r <= 0x044F {
			count += 2 // Weight lowercase more
		} else if r >= 0x0410 && r <= 0x042F {
			count += 1 // Uppercase
		}
	}
	return count
}

// DecodeFromCP1251 decodes a string that was incorrectly read as Latin-1/Windows-1252
// but was actually UTF-8 encoded. This is another common mojibake scenario.
//
// In Latin-1 and Windows-1252, the codepoint maps directly to the byte value
// for most characters (0x00-0xFF range).
func DecodeFromCP1251(mojibake string) string {
	var utf8Bytes []byte
	for _, r := range mojibake {
		// In Latin-1 and Windows-1252 (for most characters),
		// the codepoint maps directly to byte value
		if r <= 0xFF {
			utf8Bytes = append(utf8Bytes, byte(r))
		} else {
			// Keep multi-byte characters as-is
			utf8Bytes = append(utf8Bytes, []byte(string(r))...)
		}
	}
	return string(utf8Bytes)
}
