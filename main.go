package main

import (
	"BangumiBot/config"
	"BangumiBot/data"
	"fmt"
	"github.com/Logiase/gomirai"
	"github.com/Logiase/gomirai/message"
	"github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"time"
)

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
	go dailyNotifyWorker()

	// 开始监听消息
	for true {
		select {
		case <-interrupt:
			break
		case s := <-producer.Chan:
			go onPubSeason(s)
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

func dailyNotifyWorker() {
	clockStr := strings.Split(conf.Notify.DailyNotifyWhen, ":")
	if len(clockStr) < 2 {
		panic("illegal daily_notify_when value: " + conf.Notify.DailyNotifyWhen)
	}

	var err error
	clock := make([]int, len(clockStr))
	for i := range clockStr {
		clock[i], err = strconv.Atoi(clockStr[i])
		if err != nil {
			panic("illegal daily_notify_when value: " + conf.Notify.DailyNotifyWhen)
		}
	}

	if len(clock) == 2 {
		clock = append(clock, 0)
	}

	for {
		now := time.Now()
		y, m, d := now.Date()
		next := time.Date(y, m, d, clock[0], clock[1], clock[2], 0, now.Location())
		if next.Before(now) {
			next = next.AddDate(0, 0, 1)
		}

		sleepDuration := time.Duration(next.Unix()-now.Unix()) * time.Second
		logrus.Infof("next daily notify will be at %v (after %ds)", next, int64(sleepDuration.Seconds()))
		time.Sleep(sleepDuration)
		go onDailyNotify()
	}
}

func getTodaySeasonsMessage() string {
	y, m, d := time.Now().Date()
	seasons := make([]data.Season, 0)
	for _, s := range producer.Seasons() {
		y2, m2, d2 := s.PubTime.Date()
		if y == y2 && m == m2 && d == d2 {
			seasons = append(seasons, s)
		}
	}

	return temp.QueryReply(seasons)
}

func onReceiveMessage(e message.Event) {
	if len(e.MessageChain) <= 1 || e.MessageChain[1].Text != conf.General.Trigger {
		return
	}

	logrus.Infof("onTriggered")
	s := getTodaySeasonsMessage()
	reply(e, message.PlainMessage(s))
}

func onDailyNotify() {
	logrus.Info("onDailyNotify")
	msg := message.PlainMessage(getTodaySeasonsMessage())

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

func onPubSeason(s data.Season) {
	logrus.Infof("onPubSeason %v", s)
	msg := message.PlainMessage(temp.PubNotice(s))

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
