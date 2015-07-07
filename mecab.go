package main

// #cgo CFLAGS: -I/usr/include
// #cgo LDFLAGS: -L/usr/lib -lmecab -lstdc++
// #include <mecab.h>
// #include <stdio.h>
import "C"

import (
	"log"
	"strings"
	"unicode"
)

type mecabRequest struct {
	Sentence string
	Result   chan []mecabResult
}

type posType string

const (
	ADJECTIVE     posType = "形容詞"
	PRENOUN       posType = "連体詞"
	ADVERB        posType = "副詞"
	AUXILIARYVERB posType = "助動詞"
	CONJUNCTION   posType = "接続詞"
	FILLER        posType = "フィラー"
	INTERJECTION  posType = "感動詞"
	NOUN          posType = "名詞"
	OTHER         posType = "その他"
	PARTICLE      posType = "助詞"
	PREFIX        posType = "接頭詞"
	SYMBOL        posType = "記号"
	VERB          posType = "動詞"
)

// The result mecab returns
type mecabResult struct {
	Surface string  // Surface
	Pos     posType // Part of Speech
	Pos1    string  // Part of Speech 1
	Pos2    string  // Part of Speech 2
	Pos3    string  // Part of Speech 3
	ConForm string  // Conjungation Form
	ConType string  // Conjungation Type
	Base    string  // Base Form
	Read    string  // Reading
	Pron    string  // Pronounciation
}

// Parses a sentence using mecab
func mecabParser(q chan mecabRequest) {
	model := C.mecab_model_new2(C.CString(""))
	if model == nil {
		log.Panicln("Can't create a mecab model")
	}
	defer C.mecab_model_destroy(model)

	mecab := C.mecab_model_new_tagger(model)
	if mecab == nil {
		log.Panicln("Can't create a mecab tagger")
	}
	defer C.mecab_destroy(mecab)

	lattice := C.mecab_model_new_lattice(model)
	if lattice == nil {
		log.Panicln("Can't create a mecab lattice")
	}
	defer C.mecab_lattice_destroy(lattice)

	for {
		req := <-q

		res := make([]mecabResult, 0)
		C.mecab_lattice_set_sentence(lattice, C.CString(req.Sentence))
		C.mecab_parse_lattice(mecab, lattice)

		lines := strings.Split(C.GoString(C.mecab_lattice_tostr(lattice)), "\n")
		for _, l := range lines {
			if strings.Index(l, "EOS") != 0 {
				if len(l) > 1 {
					res = append(res, split(l))
				}
			}
		}

		req.Result <- res
	}
}

func split(line string) (r mecabResult) {
	var hiraKataCase = unicode.SpecialCase{
		unicode.CaseRange{0x3041, 0x3096, [unicode.MaxCase]rune{0, 0x30a1 - 0x3041, 0}},
		unicode.CaseRange{0x30a1, 0x30fa, [unicode.MaxCase]rune{0x3041 - 0x30a1, 0, 0x3041 - 0x30a1}},
	}

	l := strings.Split(line, "\t")
	r.Surface = l[0]

	features := strings.Split(l[1], ",")
	r.Pos = (posType)(features[0])
	r.Pos1 = features[1]
	r.Pos2 = features[2]
	r.Pos3 = features[3]
	r.ConForm = features[4]
	r.ConType = features[5]
	r.Base = features[6]
	if len(features) > 7 {
		r.Read = strings.ToUpperSpecial(hiraKataCase, features[7])
		r.Pron = strings.ToUpperSpecial(hiraKataCase, features[8])
	}

	return r
}
