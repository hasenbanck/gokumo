package main

import (
	"fmt"
	"gopkg.in/mgo.v2/bson"
	"html"
	"strings"
	"unicode/utf8"
)

type translationResult struct {
	Original     string
	Reading      string
	Base         string
	Translations []string
}

type queryResult struct {
	Ruby    string
	Results []translationResult
}

// TODO rename me
type oorPair struct {
	Org   string
	Orth  string
	Read  string
	Count int
}

// TODO rename me
type orPair struct {
	Orth string
	Read string
}

func getTranslations(meResults []mecabResult) []translationResult {
	trans := make([]translationResult, 0, len(meResults))
	col := session.DB("wadoku").C("dictionary")
	max := len(meResults)
	const lookforward = 6

	for i := 0; i < len(meResults); {
		r := meResults[i]
		if (r.Pos == NOUN && r.Base != "*") || r.Pos == ADJECTIVE || r.Pos == VERB {
			var rest []mecabResult
			if i+lookforward < max {
				rest = meResults[i : i+lookforward]

			} else {
				rest = meResults[i:]
			}
			pairs := make([]oorPair, 0, len(rest))
			for j := 0; j < len(rest); j++ {
				end := len(rest) - j
				var orth, read, org string
				var count int
				for k := 0; k < end; k++ {
					// Use the base / base reading of the last element
					if k+1 == end {
						orth += rest[k].Base
						read += getBaseReading(&rest[k])
					} else {
						orth += rest[k].Surface
						read += rest[k].Read
					}
					org += rest[k].Surface
					count += 1
				}
				pairs = append(pairs, oorPair{Org: org, Orth: orth, Read: read, Count: count})
			}
			// build mongo query with pairs
			q := make([]bson.M, 0, len(pairs))
			for _, t := range pairs {
				q = append(q, bson.M{"orthography": t.Orth, "reading.hiragana": t.Read})
			}

			var entry Entry
			if err := col.Find(bson.M{"$or": q}).Sort("-count").One(&entry); err == nil {
				found := false
				for _, t := range pairs {
					for _, eo := range entry.Orthography {
						if t.Orth == eo {
							found = true
							// We don't want hiragana entries, since there are too many homonyms
							if containsKanji(t.Org) || containsKatakana(t.Org) {
								trans = append(trans, translationResult{Original: t.Org, Reading: t.Read, Base: t.Orth, Translations: entry.Translation})
							}
							i += t.Count
						}
					}
				}
				if !found {
					i += 1
				}
			} else {
				i += 1
			}
		} else {
			i += 1
		}
	}
	return trans
}

func getBaseReading(meResult *mecabResult) (baseReading string) {
	surfaceFurigana := getFurigana(getGroups(meResult.Surface), meResult.Read)
	baseGroups := getGroups(meResult.Base)

	// This could result in a bug where we have the same kanji with different readings
	for _, b := range baseGroups {
		found := false
		for _, s := range surfaceFurigana {
			if b == s.Orth {
				baseReading += s.Read
				found = true
				break
			}
		}
		if !found {
			baseReading += b
		}
	}
	return
}

func getRuby(meResults []mecabResult) (ruby string) {
	for _, r := range meResults {
		// Normalization to hiragana if katakana
		surface := strings.ToUpperSpecial(hiraKataCase, r.Surface)

		// Find the kanji and create their HTML ruby
		if containsKanji(surface) {
			// First we need to "split" the surface to find the changing kanji/kana substrings
			groups := getGroups(surface)
			// Now that we know the groups, we can iter through them and create the ruby for the kanji entries
			pairs := getFurigana(groups, r.Read)
			for _, p := range pairs {
				if p.Orth == p.Read {
					ruby += p.Orth
				} else {
					ruby += fmt.Sprintf("<ruby>%s<rt>%s</rt></ruby>", p.Orth, p.Read)
				}
			}
		} else {
			ruby += r.Surface
		}
	}
	return ruby
}

// Split a string to find the changing kanji/kana substrings
func getGroups(s string) []string {
	groups := make([]string, 0, utf8.RuneCountInString(s))
	current := false
	str := ""
	for i, g := range s {
		if i == 0 {
			current = isKanji(g)
			str += fmt.Sprintf("%c", g)
		} else {
			if current != isKanji(g) {
				current = isKanji(g)
				groups = append(groups, str)
				str = fmt.Sprintf("%c", g)
			} else {
				str += fmt.Sprintf("%c", g)
			}
		}
	}
	groups = append(groups, str)
	return groups
}

func getFurigana(groups []string, reading string) []orPair {
	pairs := make([]orPair, 0, len(groups))
	for i, g := range groups {
		if strings.HasPrefix(reading, g) {
			reading = strings.TrimPrefix(reading, g)
			pairs = append(pairs, orPair{Orth: g, Read: g})
		} else {
			if i+1 < len(groups) {
				next := groups[i+1]
				spl := strings.SplitN(reading, next, 2)
				if len(spl) > 0 {
					reading = strings.TrimPrefix(reading, spl[0])
					pairs = append(pairs, orPair{Orth: g, Read: spl[0]})
				}
			} else {
				pairs = append(pairs, orPair{Orth: g, Read: reading})
			}
		}
	}
	return pairs
}

func queryMecab(query string) []mecabResult {
	// Prepare the request for mecab
	var mr mecabRequest
	mr.Sentence = html.EscapeString(query)
	mr.Result = make(chan []mecabResult, 2)

	// Query mecab over a channel
	mc <- mr
	return <-mr.Result
}
