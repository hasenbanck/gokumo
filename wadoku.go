package main

type grammType int

const (
	NotDefined grammType = iota
	Daimeishi
	Doushi
	Fukujoshi
	Fukushi
	Jodoushi
	Jokeiyoushi
	Joshi
	Kakarijoshi
	Kandoushi
	Kanji
	Keiyoudoushi
	Keiyoushi
	Meishi
	Prefix
	Rengo
	Rentaishi
	Setsuzokushi
	Shuujoshi
	Specialcharacter
	Suffix
	Wordcomponent
)

type Reading struct {
	Hiragana string
	Hatsuon  string
}

type Entry struct {
	Id          int
	GrammType   grammType
	Orthography []string
	Reading     []Reading
	Translation []string
	Count       int
}
