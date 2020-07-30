package main

import (
	"BangumiBot/config"
	"BangumiBot/data"
	"BangumiBot/templater"
	"fmt"
	"github.com/Logiase/gomirai"
	"github.com/Logiase/gomirai/message"
	"github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"time"
)

var client *gomirai.Client
var bot *gomirai.Bot
var conf config.Config
var producer = data.NewSeasonProducer()

func main() {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	var err error
	conf, err = config.LoadConfig("config.json")
	if err != nil {
		logrus.Error(err)
	}

	// 初始化Bot部分
	url := fmt.Sprintf("http://%s:%d", conf.Mirai.Host, conf.Mirai.Port)
	client = gomirai.NewClient("default", url, conf.Mirai.AuthKey)
	tryAuth()
	defer shutdownBot()

	// 启动一个goroutine用于接收消息
	go func() {
		for {
			err := bot.FetchMessages()
			if err != nil {
				client.Logger.Error(err)
				tryAuth()
			}
		}
	}()

	producer.Start(time.Duration(conf.General.FetchDuration) * time.Second)

	// 开始监听消息
	for true {
		select {
		case <-interrupt:
			break
		case s := <-producer.Chan:
			onPubSeason(s)
		case e := <-bot.Chan:
			switch e.Type {
			case message.EventReceiveFriendMessage,
				message.EventReceiveGroupMessage,
				message.EventReceiveTempMessage:
				go onReceiveMessage(e)
			}
		}
	}
}

func onReceiveMessage(e message.Event) {
	if len(e.MessageChain) <= 1 || e.MessageChain[1].Text != conf.General.Trigger {
		return
	}

	y, m, d := time.Now().Date()
	seasons := make([]data.Season, 0)
	for _, s := range producer.Seasons() {
		y2, m2, d2 := s.PubTime.Date()
		if y == y2 && m == m2 && d == d2 {
			seasons = append(seasons, s)
		}
	}

	reply(e, message.PlainMessage(templater.QueryReply(seasons)))
}

func onPubSeason(s data.Season) {
	msg := message.PlainMessage(templater.PubNotice(s))

	for _, fid := range conf.Notify.Friend {
		_, err := bot.SendFriendMessage(fid, 0, msg)
		if err != nil {
			client.Logger.Error(err)
		}
	}

	for _, gid := range conf.Notify.Group {
		_, err := bot.SendGroupMessage(gid, 0, msg)
		if err != nil {
			client.Logger.Error(err)
		}
	}
}
