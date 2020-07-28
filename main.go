package main

import (
	"BangumiBot/data"
	"BangumiBot/templater"
	"encoding/json"
	"fmt"
	"github.com/Logiase/gomirai"
	"github.com/Logiase/gomirai/message"
	"io/ioutil"
	"os"
	"os/signal"
	"time"
)

var client *gomirai.Client
var bot *gomirai.Bot
var config Config
var producer = data.NewSeasonProducer()

func loadConfig() error {
	b, err := ioutil.ReadFile("config.json")
	if err != nil {
		if os.IsNotExist(err) {
			config = DefaultConfig
			b, err = json.Marshal(DefaultConfig)
			if err != nil {
				return err
			}
			err = ioutil.WriteFile("config.json", b, 0644)
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}
	return json.Unmarshal(b, &config)
}

func main() {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	go func() {
		<-interrupt
		onExit()
		os.Exit(0)
	}()

	err := loadConfig()
	if err != nil {
		fmt.Println(err)
	}

	// 初始化Bot部分
	url := fmt.Sprintf("http://%s:%d", config.Mirai.Host, config.Mirai.Port)
	client = gomirai.NewClient("default", url, config.Mirai.AuthKey)
	tryAuth()

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

	producer.Start(config.General.FetchDuration * time.Second)

	// 开始监听消息
	for true {
		select {
		case s := <-producer.Chan:
			onPubSeason(s)
		case e := <-bot.Chan:
			switch e.Type {
			case message.EventReceiveFriendMessage, message.EventReceiveGroupMessage, message.EventReceiveTempMessage:
				go onReceiveMessage(e)
			}
		}
	}
}

func onExit() {
	err := client.Release(config.Mirai.QQ)
	if err != nil {
		client.Logger.Warn(err)
	}
}

func onReceiveMessage(e message.Event) {
	if e.MessageChain[1].Text != config.General.Trigger {
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

	for _, fid := range config.Notify.Friend {
		_, err := bot.SendFriendMessage(fid, 0, msg)
		if err != nil {
			client.Logger.Error(err)
		}
	}

	for _, gid := range config.Notify.Group {
		_, err := bot.SendGroupMessage(gid, 0, msg)
		if err != nil {
			client.Logger.Error(err)
		}
	}
}
