package main

import "time"

type Config struct {
	Mirai   MiraiConfig   `json:"mirai"`
	General GeneralConfig `json:"general"`
	Notify  NotifyConfig  `json:"notify"`
}

type MiraiConfig struct {
	QQ      uint   `json:"qq"`
	Host    string `json:"host"`
	Port    uint   `json:"port"`
	AuthKey string `json:"auth_key"`
}

type GeneralConfig struct {
	Trigger       string        `json:"trigger"`
	FetchDuration time.Duration `json:"fetch_duration"`
}

type NotifyConfig struct {
	Friend []uint `json:"friend"`
	Group  []uint `json:"group"`
}

var DefaultConfig = Config{
	Mirai: MiraiConfig{
		QQ:      123456,
		Host:    "localhost",
		Port:    3399,
		AuthKey: "qwerty",
	},
}
