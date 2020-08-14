package main

import (
	"BangumiBot/config"
	"BangumiBot/data"
	"BangumiBot/templater"
	"github.com/Logiase/gomirai"
)

var client *gomirai.Client
var bot *gomirai.Bot
var conf config.Config

var producer = data.NewSeasonProducer()
var temp = templater.LoadGlob("template/*")
