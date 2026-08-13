package telegramapi

import "unicode/utf16"

// UTF16Len returns the number of UTF-16 code units Telegram uses for entity
// offsets, entity lengths, and message length limits.
func UTF16Len(text string) int {
	units := 0
	for _, r := range text {
		units++
		if r > 0xffff {
			units++
		}
	}
	return units
}

// SliceUTF16 extracts the exact Telegram entity range. It rejects offsets that
// split a surrogate pair instead of silently returning the wrong text.
func SliceUTF16(text string, offset, length int) (string, bool) {
	if offset < 0 || length < 0 {
		return "", false
	}
	units := utf16.Encode([]rune(text))
	end := offset + length
	if offset > len(units) || end < offset || end > len(units) {
		return "", false
	}
	if offset > 0 && offset < len(units) && isLowSurrogate(units[offset]) {
		return "", false
	}
	if end > 0 && end < len(units) && isLowSurrogate(units[end]) {
		return "", false
	}
	return string(utf16.Decode(units[offset:end])), true
}

// SplitUTF16 takes at most limit Telegram length units from the front without
// splitting a Unicode code point.
func SplitUTF16(text string, limit int) (head, tail string) {
	if limit <= 0 {
		return "", text
	}
	units := 0
	byteAt := 0
	for at, r := range text {
		width := 1
		if r > 0xffff {
			width = 2
		}
		if units+width > limit {
			return text[:at], text[at:]
		}
		units += width
		byteAt = at + len(string(r))
	}
	return text[:byteAt], text[byteAt:]
}

func isLowSurrogate(unit uint16) bool { return unit >= 0xdc00 && unit <= 0xdfff }
