package main

import "unicode"

var hiraKataCase = unicode.SpecialCase{
	unicode.CaseRange{0x3041, 0x3096, [unicode.MaxCase]rune{0, 0x30a1 - 0x3041, 0}},
	unicode.CaseRange{0x30a1, 0x30fa, [unicode.MaxCase]rune{0x3041 - 0x30a1, 0, 0x3041 - 0x30a1}},
}

func isKatakana(r rune) bool {
	return unicode.IsOneOf([]*unicode.RangeTable{unicode.Katakana}, r)
}

func isKanji(r rune) bool {
	return unicode.IsOneOf([]*unicode.RangeTable{unicode.Ideographic}, r)
}

func containsKanji(s string) bool {
	for _, r := range s {
		if isKanji(r) {
			return true
		}
	}
	return false
}

func containsKatakana(s string) bool {
	for _, r := range s {
		if isKatakana(r) {
			return true
		}
	}
	return false
}
