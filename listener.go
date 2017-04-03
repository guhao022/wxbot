package main

import (
	"axiom"
	"encoding/json"
	"github.com/FrankWong1213/golang-lunar"
	"net/http"
	"net/url"
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
		}, {
			Regex: "明天农历|明天是农历|农历明天",
			HandlerFunc: func(c *axiom.Context) {
				t := time.Now().AddDate(0, 0, 1)
				c.Reply(" 明天是农历 " + lunar.Lunar(t.Format("20060102")))
			},
		}, /*{
			Regex: "",
			HandlerFunc: func(c *axiom.Context) {
				msg, err := w.tuling(c.Message.Text, c.Message.ToAXID)
				if err != nil {
					msg = "你继续说..."
				}
				c.Reply(msg)
			},
		},*/
	}
}

func (w *WeChatListener) tuling(msg, uid string) (string, error) {
	key := "94e05d8c1378437ebccbd735f1944755"

	var remsg = ""

	values := url.Values{}

	values.Add(`key`, key)
	values.Add(`info`, msg)
	values.Add(`userid`, uid)

	resp, err := http.PostForm(`http://www.tuling123.com/openapi/api`, values)
	if err != nil {
		return remsg, err
	}

	reader := resp.Body
	defer resp.Body.Close()

	result := make(map[string]interface{})

	err = json.NewDecoder(reader).Decode(&result)
	if err != nil {
		return remsg, err
	}

	code := result[`code`].(float64)

	if code == 100000 {
		remsg = result[`text`].(string)
	} else if code == 200000 {
		text := result[`text`].(string)
		url := result[`url`].(string)
		remsg = text + `
		` + url
	}

	return remsg, nil

}
