package main

import (
	"axiom"
	"time"
	"wxbot/tools/times"
)

type WeChatListener struct{}

func (w *WeChatListener) Handle() []*axiom.Listener {

	return []*axiom.Listener{
		{
			Regex: "time|时间|几点",
			HandlerFunc: func(c *axiom.Context) {
				layout := "2006-01-02 15:04:05"
				t := time.Now()
				c.Reply(" 现在时间: " + t.Format(layout))
			},
		}, {
			Regex: "今天星期几|今天周几",
			HandlerFunc: func(c *axiom.Context) {
				t := time.Now()
				c.Reply(" 今天" + times.WeekdayText(t.Weekday().String()) + "了")
			},
		},{
			Regex: "明天星期几|明天周几",
			HandlerFunc: func(c *axiom.Context) {
				t := time.Now().AddDate(0,0,1)
				c.Reply(" 明天是" + times.WeekdayText(t.Weekday().String()))
			},
		},
	}
}

