package main

import (
	"sync"
	"wxbot/wechat"
	"wxbot/tools/log"
)

type xiaoice struct {
	sync.Mutex
	un       string
	waitting []string
	bot      *WeChat
}

func newXiaoice(wx *WeChat) *xiaoice {
	x := &xiaoice{}
	x.bot = wx
	return x
}

func (x *xiaoice) autoReplay(msg wechat.EventMsgData) {
	if msg.IsSendedByMySelf {
		return
	}

	if msg.FromUserName == x.un { // 小冰发来的消息
		x.Lock()
		x.Unlock()

		count := len(x.waitting)
		if count == 0 {
			log.Warnf(`msg Form xiaoice %s`, msg.Content)
			return
		}
		to := x.waitting[count-1]
		x.waitting = x.waitting[:count-1]

		if msg.IsMediaMsg {
			if path, err := x.bot.Wechat.DownloadMedia(msg.MediaURL, msg.OriginalMsg[`MsgId`].(string)); err == nil {
				x.bot.Wechat.SendFile(path, to)
			}
		} else {
			x.bot.Wechat.SendTextMsg(msg.Content, to)
		}
	} else if !msg.IsSendedByMySelf { // 转发别人的消息到小冰
		var err error
		if msg.IsMediaMsg {
			if path, e := x.bot.Wechat.DownloadMedia(msg.MediaURL, msg.OriginalMsg[`MsgId`].(string)); e == nil {
				err = x.bot.Wechat.SendFile(path, x.un)
			} else {
				err = e
			}
		} else {
			println(msg.Content)
			err = x.bot.Wechat.SendTextMsg(msg.Content, x.un)
		}

		if err == nil {
			x.Lock()
			defer x.Unlock()
			x.waitting = append(x.waitting, msg.FromUserName)
		} else {
			log.Error(err)
		}
	}
}
