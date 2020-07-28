package data

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

type RemoteResult struct {
	Code    int          `json:"code"`
	Message string       `json:"message"`
	Result  []DaySeasons `json:"result"`
}

type DaySeasons struct {
	DateStr   string `json:"date"`
	Date      time.Time
	DateTs    int64    `json:"date_ts"`
	DayOfWeek int      `json:"day_of_week"`
	IsToday   int      `json:"is_today"`
	Seasons   []Season `json:"seasons"`
}

type Season struct {
	Cover        string `json:"cover"`
	Delay        int    `json:"delay"`
	EpId         int64  `json:"ep_id"`
	Favorites    int64  `json:"favorites"`
	Follow       int    `json:"follow"`
	IsPublished  int    `json:"is_published"`
	PubIndex     string `json:"pub_index"`
	PubTimeStr   string `json:"pub_time"`
	PubTime      time.Time
	PubTs        int64  `json:"pub_ts"`
	SeasonId     int64  `json:"season_id"`
	SeasonStatus int    `json:"season_status"`
	SquareCover  string `json:"square_cover"`
	Title        string `json:"title"`
	Url          string `json:"url"`
}

func (s Season) String() string {
	return fmt.Sprintf("(%s) %s - %s",
		s.PubTime.Format("1-2 15:04"),
		s.Title, s.PubIndex)
}

func FetchRemote() ([]Season, error) {
	res, err := http.Get("https://bangumi.bilibili.com/web_api/timeline_global")
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var remoteRes RemoteResult
	err = json.Unmarshal(body, &remoteRes)
	if err != nil {
		return nil, err
	}
	if remoteRes.Code != 0 {
		return nil, fmt.Errorf("fetched fail code: %d", remoteRes.Code)
	}

	// 写入PubTime
	for i, d := range remoteRes.Result {
		for j, s := range d.Seasons {
			remoteRes.Result[i].Seasons[j].PubTime = time.Unix(s.PubTs, 0)
		}
	}

	seasons := make([]Season, 0)
	for _, day := range remoteRes.Result {
		seasons = append(seasons, day.Seasons...)
	}

	return seasons, nil
}
