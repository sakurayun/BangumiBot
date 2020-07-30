package data

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"time"
)

type Query struct {
	Type   string
	Data   interface{}
	Result chan<- interface{}
}

const (
	qGetAll         = "get_all"
	qGetNextPending = "get_next_pending"
	qUpdate         = "update"
	qDeliver        = "qDeliver"
)

type SeasonProducer struct {
	workQueue chan Query

	fetchRemote func() ([]Season, error)

	Chan    chan Season
	all     []Season
	pending []Season

	updateEvent chan struct{}
}

func NewSeasonProducer() *SeasonProducer {
	return &SeasonProducer{
		workQueue:   make(chan Query, 16),
		fetchRemote: FetchRemote,
		Chan:        make(chan Season, 16),
		all:         []Season{},
		pending:     []Season{},
		updateEvent: make(chan struct{}),
	}
}

// 管理all和pending字段
func (p *SeasonProducer) manager() {
	for q := range p.workQueue {
		switch q.Type {
		case qGetAll: // 获取所有番剧
			q.Result <- p.all
		case qGetNextPending: // 获取下一个番剧
			if len(p.pending) == 0 {
				q.Result <- nil
			} else {
				q.Result <- p.pending[0]
			}
		case qUpdate: // 更新数据
			data := q.Data.([]Season)
			p.doUpdate(data)
			q.Result <- nil
		case qDeliver: // 马上分发下一个番剧
			err := p.doDeliver()
			q.Result <- err
		}
	}
}

func (p *SeasonProducer) doUpdate(data []Season) {
	p.all = data

	if len(p.all) > 0 {
		// 二分查找第一个PubTime不先于当前时间的位置
		now := time.Now()

		l, r := 0, len(p.all)
		for l < r {
			mid := (l + r) / 2
			if p.all[mid].PubTime.Before(now) {
				l = mid + 1
			} else {
				r = mid
			}
		}

		p.pending = p.all[l:]
	} else {
		p.pending = p.all
	}
	p.updateEvent <- struct{}{}
	logrus.Infof("update: %d seasons, %d are pending", len(p.all), len(p.pending))
}

func (p *SeasonProducer) doDeliver() error {
	if len(p.pending) == 0 {
		return fmt.Errorf("no pending season")
	}
	head := p.pending[0]
	p.Chan <- head
	p.pending = p.pending[1:]
	logrus.Infof("deliver: %s - %s", head.Title, head.PubIndex)
	return nil
}

func (p *SeasonProducer) Start(fetchDuration time.Duration) {
	go p.manager()
	go p.poster()
	go p.updater(fetchDuration)
}

// 获取所有番剧
func (p *SeasonProducer) Seasons() []Season {
	ch := make(chan interface{})
	p.workQueue <- Query{qGetAll, nil, ch}
	return (<-ch).([]Season)
}

// 按时获取下一个番剧并分发
func (p *SeasonProducer) poster() {
	wait := time.After(time.Hour * 24 * 7) // poster: first i want to delay a very very long duration
	for true {
		select {
		case <-wait: // poster: it's time to deliver the next pending season
			ch := make(chan interface{})
			p.workQueue <- Query{qDeliver, nil, ch}
			if err := <-ch; err != error(nil) {
				logrus.Error(err)
			}
		case <-p.updateEvent: // manager: wake up now to see the next pending season
		}

		// poster: now i should look up the next pending season
		ch := make(chan interface{})
		p.workQueue <- Query{qGetNextPending, nil, ch}
		if head, ok := (<-ch).(Season); ok {
			wait = time.After(time.Second * time.Duration(head.PubTime.Unix()-time.Now().Unix()))
		} else {
			wait = time.After(time.Hour * 24 * 7)
		}
	}
}

// 每隔2h从bilibili偷一次时间表
func (p *SeasonProducer) updater(fetchDuration time.Duration) {
	for true {
		s, err := p.fetchRemote()
		if err != nil {
			logrus.Error(err)
		}

		ch := make(chan interface{})
		q := Query{qUpdate, s, ch}
		p.workQueue <- q
		<-ch

		time.Sleep(fetchDuration)
	}
}
