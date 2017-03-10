package main

import (
	"axiom"
	"wxbot/wechat"
	"fmt"
)

type WeChat struct {
	bot *axiom.Robot
	wechat *wechat.WeChat
}

func NewWeChat(bot *axiom.Robot) *WeChat {

	options := wechat.DefaultConfigure()

	wechat, err := wechat.AwakenNewBot(options)
	if err != nil {
		panic(err)
	}

	return &WeChat{
		bot: bot,
		wechat: wechat,
	}
}

// 初始化
func (w *WeChat) Construct() error {

	w.wechat.Go()

	w.wechat.Handle(`/login`, func(arg wechat.Event) {
		isSuccess := arg.Data.(int) == 1
		if isSuccess {
			fmt.Println(`login Success`)
		} else {
			fmt.Println(`login Failed`)
		}
	})

	return nil
}

// 解析
func (w *WeChat) Process() error {

	return nil
}

// 回应
func (w *WeChat) Reply(msg axiom.Message, message string) error {

	return nil
}

