package main

import (
	"github.com/Logiase/gomirai/message"
	"time"
)

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
			client.Logger.Info("will retry to auth in 5 secs")
			<-ticker
		}
	}
	client.Logger.Info("authed successfully")
}

func reply(e message.Event, msg ...message.Message) {
	src := e.MessageChain[0].Id

	var err error
	switch e.Type {
	case message.EventReceiveFriendMessage:
		_, err = bot.SendFriendMessage(e.Sender.Id, src, msg...)
	case message.EventReceiveGroupMessage:
		_, err = bot.SendGroupMessage(e.Sender.Group.Id, src, msg...)
	case message.EventReceiveTempMessage:
		_, err = bot.SendTempMessage(e.Sender.Group.Id, e.Sender.Id, msg...)
	}

	if err != nil {
		client.Logger.Error(err)
	}
}
