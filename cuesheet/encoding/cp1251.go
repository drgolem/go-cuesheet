// Package encoding provides character encoding utilities for CUE sheet processing.
// It delegates to github.com/drgolem/cyrillic-encoding for the core implementation.
package encoding

import cyrillic "github.com/drgolem/cyrillic-encoding"

// CP1251ToByte converts a Unicode character to its CP1251 (Windows-1251) byte value.
// Returns 0 if the character is not in CP1251 encoding (except for actual 0x00).
func CP1251ToByte(r rune) byte {
	return cyrillic.CP1251ToByte(r)
}

// DecodeMojibakeFromCP1251 fixes UTF-8 text that was incorrectly read as CP1251 (Windows-1251).
func DecodeMojibakeFromCP1251(mojibake string) string {
	return cyrillic.DecodeMojibakeFromCP1251(mojibake)
}

// CountCyrillic counts the number of Cyrillic characters in a string.
// Lowercase letters are weighted more heavily (2x) as they're more common in normal text.
func CountCyrillic(s string) int {
	return cyrillic.CountCyrillic(s)
}

// DecodeFromCP1251 decodes a string that was incorrectly read as Latin-1/Windows-1252
// but was actually UTF-8 encoded.
//
// Deprecated: Use cyrillic.DecodeMojibakeFromISO8859 directly.
func DecodeFromCP1251(mojibake string) string {
	return cyrillic.DecodeMojibakeFromISO8859(mojibake)
}
