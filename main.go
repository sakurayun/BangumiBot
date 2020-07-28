package main

import (
	"BangumiBot/data"
	"bytes"
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

func tryAuth() {
	ticker := time.Tick(5 * time.Second)

	for {
		ok := true

		key, err := client.Auth()
		if err != nil {
			client.Logger.Error(err)
			ok = false
		}
		bot, err = client.Verify(config.Mirai.QQ, key)
		if err != nil {
			client.Logger.Error(err)
			ok = false
		}

		if ok {
			break
		} else {
			<-ticker
		}
	}
}

func onExit() {
	err := client.Release(config.Mirai.QQ)
	if err != nil {
		client.Logger.Warn(err)
	}
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

func reply(e message.Event, msg ...message.Message) {
	src := e.MessageChain[0].Id

	var err error
	switch e.Type {
	case message.EventReceiveFriendMessage:
		_, err = bot.SendFriendMessage(e.Sender.Id, src, msg...)
	case message.EventReceiveGroupMessage:
		_, err = bot.SendGroupMessage(e.Sender.Group.Id, src, msg...)
	}

	if err != nil {
		client.Logger.Error(err)
	}
}

func onReceiveMessage(e message.Event) {
	if e.MessageChain[1].Text != config.General.Trigger {
		return
	}

	now := time.Now()

	buffer := bytes.Buffer{}
	for _, s := range producer.Seasons() {
		y, m, d := s.PubTime.Date()
		y2, m2, d2 := now.Date()
		if y == y2 && m == m2 && d == d2 {
			buffer.WriteString(s.String())
			buffer.WriteRune('\n')
		}
	}

	reply(e, message.PlainMessage(buffer.String()))
}

func onPubSeason(s data.Season) {
	msg := message.PlainMessage(s.String())

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
