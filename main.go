package main

import (
	"github.com/gin-gonic/gin"
	"gopkg.in/mgo.v2"
	"log"
	"net/http"
	"runtime"
)

var session *mgo.Session
var mc chan mecabRequest

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	router := gin.Default()
	router.POST("/query", queryController)
	router.StaticFile("/", "./static/index.html")
	router.StaticFile("/favicon.ico", "./static/favicon.ico")
	router.Static("/css", "./static/css")
	router.Static("/fonts", "./static/fonts")
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

func queryController(c *gin.Context) {
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
