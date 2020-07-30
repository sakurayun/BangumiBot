package config

import (
	"encoding/json"
	"io/ioutil"
	"math/rand"
	"os"
	"reflect"
	"strconv"
	"testing"
)

func isFileExists(t *testing.T, filename string) bool {
	t.Helper()
	_, err := os.Stat(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return false
		} else {
			panic(err)
		}
	}
	return true
}

func withTempFile(t *testing.T, action func(t *testing.T, filename string)) {
	t.Helper()
	filename := "config_test" + strconv.FormatUint(rand.Uint64(), 10) + ".json"
	if isFileExists(t, filename) {
		_ = os.Remove(filename)
	}
	defer os.Remove(filename)

	action(t, filename)
}

const goodConfigFile = `{
  "mirai": {
    "qq": 123456,
    "host": "localhost",
    "port": 3399,
    "auth_key": "qwerty"
  },
  "general": {
    "trigger": "看看番",
    "fetch_duration": 7200
  },
  "notify": {
    "friend": [114514],
    "group": [1919810]
  }
}`

var goodConfig = Config{
	Mirai: MiraiConfig{
		QQ:      123456,
		Host:    "localhost",
		Port:    3399,
		AuthKey: "qwerty",
	},
	General: GeneralConfig{
		Trigger:       "看看番",
		FetchDuration: 7200,
	},
	Notify: NotifyConfig{
		Friend: []uint{114514},
		Group:  []uint{1919810},
	},
}

func TestLoadConfig(t *testing.T) {
	t.Run("no config file", func(t *testing.T) {
		withTempFile(t, func(t *testing.T, filename string) {
			conf, err := LoadConfig(filename)
			if err != nil {
				t.Error(err)
			}

			if !reflect.DeepEqual(conf, DefaultConfig) {
				t.Error("config got not equal to the default")
			}

			if !isFileExists(t, filename) {
				t.Error("config file was not generated")
			} else {
				b, err := ioutil.ReadFile(filename)
				if err != nil {
					panic(err)
				}

				err = json.Unmarshal(b, &conf)
				if err != nil {
					t.Error(err)
				}

				if !reflect.DeepEqual(conf, DefaultConfig) {
					t.Error("config loaded from file not equal to the default")
				}
			}
		})
	})

	t.Run("good config file", func(t *testing.T) {
		withTempFile(t, func(t *testing.T, filename string) {
			err := ioutil.WriteFile(filename, []byte(goodConfigFile), 0644)
			if err != nil {
				panic(err)
			}

			conf, err := LoadConfig(filename)
			if err != nil {
				t.Error(err)
			}

			if !reflect.DeepEqual(conf, goodConfig) {
				t.Errorf("got: %v\nwants: %v", conf, goodConfig)
			}
		})
	})
}
