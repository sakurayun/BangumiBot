package data

import (
	"fmt"
	"strconv"
	"testing"
	"time"
)

func TestSeasonProducer(t *testing.T) {
	p := NewSeasonProducer()
	p.fetchRemote = func() ([]Season, error) {
		seasons := make([]Season, 0)
		for i := 0; i < 10; i++ {
			s := Season{
				PubTime: time.Now().Add(time.Duration(i+2) * time.Second),
				Title:   strconv.Itoa(i),
			}
			seasons = append(seasons, s)
		}
		return seasons, nil
	}

	p.Start(time.Minute * 60)

	go func() {
		time.Sleep(20 * time.Second)
		t.Errorf("every season was not collected")
	}()

	for i := 0; i < 10; i++ {
		s := <-p.Chan
		fmt.Println("collected " + s.Title)
	}
}
