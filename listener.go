package main

import (
	"axiom"
	"github.com/FrankWong1213/golang-lunar"
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
				c.Reply(t.Format(layout))
			},
		}, {
			Regex: "今天星期几|今天周几|周几|星期|星期几",
			HandlerFunc: func(c *axiom.Context) {
				t := time.Now()
				c.Reply(" 今天" + times.WeekdayText(t.Weekday().String()) + "了")
			},
		}, {
			Regex: "明天星期几|明天周几",
			HandlerFunc: func(c *axiom.Context) {
				t := time.Now().AddDate(0, 0, 1)
				c.Reply(" 明天是" + times.WeekdayText(t.Weekday().String()))
			},
		}, {
			Regex: "今天农历|今天是农历|农历今天|农历",
			HandlerFunc: func(c *axiom.Context) {
				t := time.Now()
				c.Reply(" 今天是农历 " + lunar.Lunar(t.Format("20060102")))
			},
		},{
			Regex: "明天农历|明天是农历|农历明天",
			HandlerFunc: func(c *axiom.Context) {
				t := time.Now().AddDate(0, 0, 1)
				c.Reply(" 明天是农历 " + lunar.Lunar(t.Format("20060102")))
			},
		},
	}
}

