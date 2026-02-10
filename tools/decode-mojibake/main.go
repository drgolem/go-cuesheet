package main

import (
	"fmt"
	"os"
	"unicode/utf8"

	"github.com/drgolem/go-cuesheet/cuesheet/encoding"
)

func main() {
	fmt.Println("Mojibake Decoder - Fix double-encoded text")
	fmt.Println("Supports: UTF-8 misread as Windows-1252, Latin-1, or CP1251 (Cyrillic)")
	fmt.Println()

	if len(os.Args) < 2 {
		fmt.Println("Usage: decode-mojibake <text>")
		fmt.Println()
		fmt.Println("Examples:")
		fmt.Println("  UTF-8 misread as Windows-1252:")
		fmt.Println("    'ÐÑÐ°Ð²Ð¾' → 'Браво' (Bravo)")
		fmt.Println()
		fmt.Println("  UTF-8 misread as CP1251 (Cyrillic):")
		fmt.Println("    'Р'СЂР°РІРѕ' → 'Браво' (Bravo)")
		fmt.Println("    'РџСЂР°РІРѕ' → 'Право' (Right/Law)")
		os.Exit(1)
	}

	mojibake := os.Args[1]

	fmt.Printf("Input (mojibake): %s\n", mojibake)
	fmt.Println()

	// Try different decodings
	fmt.Println("Trying different encodings:")
	fmt.Println()

	// Try UTF-8 misread as Latin-1/Windows-1252
	if decoded := encoding.DecodeFromCP1251(mojibake); utf8.ValidString(decoded) && decoded != mojibake {
		fmt.Printf("  Latin-1/Windows-1252: %s\n", decoded)
	}

	// Try UTF-8 misread as CP1251 (common for Russian text)
	if decoded := encoding.DecodeMojibakeFromCP1251(mojibake); utf8.ValidString(decoded) && decoded != mojibake {
		fmt.Printf("  CP1251 (Cyrillic):    %s ✓\n", decoded)
	}
}

