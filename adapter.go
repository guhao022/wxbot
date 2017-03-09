package main

import (
	"axiom"
	"wxbot/wechat"
)

type WeChat struct {
	bot *axiom.Robot
}

func NewWeChat(bot *axiom.Robot) *WeChat {

	options := wechat.DefaultConfigure()

	wechat, err := wechat.AwakenNewBot(options)
	if err != nil {
		panic(err)
	}

	wechat.Handle()

	return &WeChat{
		bot: bot,
	}
}

// 初始化
func (w *WeChat) Construct() error {

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

