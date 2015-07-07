package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"html"
	"log"
	"net/http"
	"runtime"
	"strings"
)

type translationResult struct {
	Original     string
	Base         string
	Translations []string
}

type queryResult struct {
	Ruby    string
	Results []translationResult
}

var session *mgo.Session
var mc chan mecabRequest

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	router := gin.Default()
	router.POST("/query", query)
	router.StaticFile("/", "./static/index.html")
	router.StaticFile("/favicon.ico", "./static/favicon.ico")
	router.Static("/css", "./static/css")
	router.Static("/js", "./static/js")

	// Mongo DB
	var err error
	if session, err = mgo.Dial("127.0.0.1"); err != nil {
		log.Fatalln("Can't connect to mongodb: " + err.Error())
	}
	defer session.Close()

	// Switch the session to a eventual behavior.
	session.SetMode(mgo.Eventual, true)

	// Initialize mecab
	mc = make(chan mecabRequest, 10)
	go mecabParser(mc)

	router.Run(":8080")
}

func query(c *gin.Context) {
	// Get the query string
	query := c.PostForm("query")
	if query != "" {
		res := queryMecab(query)

		trans := getTranslations(res)
		ruby := getRuby(res)

		c.JSON(http.StatusOK, queryResult{ruby, trans})
	} else {
		c.JSON(http.StatusOK, queryResult{"", []translationResult{}})
	}
}

func getTranslations(meResults []mecabResult) []translationResult {
	trans := make([]translationResult, 0, len(meResults))
	col := session.DB("wadoku").C("dictionary")
	for _, r := range meResults {
		// TODO compile "sane" searches bevor quering to the database
		if (r.Pos == NOUN && r.Base != "*") || r.Pos == ADJECTIVE || r.Pos == VERB {
			var entry Entry
			// We may want all results and iterate over them to find the right one (reading!)
			if err := col.Find(bson.M{"orthography": r.Base}).One(&entry); err == nil {
				trans = append(trans, translationResult{Original: r.Surface, Base: r.Base, Translations: entry.Translation})
			}
		}
	}
	return trans
}

func getRuby(meResults []mecabResult) (ruby string) {
	for _, r := range meResults {
		// Normalization to hiragana if katakana
		surface := strings.ToUpperSpecial(hiraKataCase, r.Surface)
		// Find the kanji and create their HTML ruby
		if surface != r.Pron && r.Pron != "" {
			kanji := strings.TrimRightFunc(surface, isHiragana)
			kana := strings.TrimLeftFunc(surface, isKanji)
			reading := strings.TrimRight(r.Pron, kana)
			ruby += fmt.Sprintf("<ruby>%s<rt>%s</rt>%s</ruby>", kanji, reading, kana)
		} else {
			ruby += r.Surface
		}
	}
	return ruby
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
