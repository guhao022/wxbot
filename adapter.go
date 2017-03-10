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

	wechat, err := wechat.AwakenNewBot(nil)
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

	w.wechat.Handle(`/login`, func(arg2 wechat.Event) {})

	return nil
}

// 解析
func (w *WeChat) Process() error {

	//
	w.wechat.Handle(`/msg`, func(evt wechat.Event) {
		msg := evt.Data.(wechat.EventMsgData)

		if msg.IsGroupMsg {
			w.wechat.SendTextMsg(msg.Content, msg.FromUserName)
		}

	})

	/*w.wechat.Handle(`/msg/group`, func(evt wechat.Event) {
		data := evt.Data.(wechat.EventMsgData)
		content := data.Content
		w.wechat.SendTextMsg(content, data.FromGGID)
		*//*if data.AtMe {



			*//**//*if content == "统计人数" {
				stat, err := w.webGetChatRoomMember(data.FromUserName)
				if err == nil {
					ans := "据统计群里男生" + stat["man"] + "人，女生" + stat["woman"] + "人 (ó-ò)"

					w.wechat.SendTextMsg(ans, data.FromUserName)
				}
			}*//**//*
		}*//*
	})*/

	w.wechat.Handle(`/contact`, func(evt wechat.Event) {
		data := evt.Data.(wechat.EventContactData)
		fmt.Println(`/contact` + data.GGID)
	})

	w.wechat.Go()
	return nil
}

// 回应
func (w *WeChat) Reply(msg axiom.Message, message string) error {

	return nil
}

