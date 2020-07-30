package config

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

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
	Trigger       string `json:"trigger"`
	FetchDuration uint   `json:"fetch_duration"`
}

type NotifyConfig struct {
	Friend []uint `json:"friend"`
	Group  []uint `json:"group"`
}

var DefaultConfig = Config{}

func LoadConfig(filename string) (Config, error) {
	var config Config
	var err error

	b, err := ioutil.ReadFile(filename)

	if err != nil {
		if os.IsNotExist(err) {
			// 文件不存在，将写入默认配置
			config = DefaultConfig

			b, err = json.Marshal(DefaultConfig)
			if err != nil {
				return config, err
			}

			err = ioutil.WriteFile(filename, b, 0644)
			if err != nil {
				return config, err
			}
			return config, nil
		} else {
			return config, err
		}
	}

	err = json.Unmarshal(b, &config)
	if err != nil {
		return config, err
	}
	return config, nil
}
