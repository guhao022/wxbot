package main

import (
	"axiom"
	"strings"
	"wxbot/wechat"
	"fmt"
)

type WeChat struct {
	bot    *axiom.Robot
	Wechat *wechat.WeChat
}

func NewWeChat(bot *axiom.Robot) *WeChat {

	wechat, err := wechat.AwakenNewBot(nil)
	if err != nil {
		panic(err)
	}

	return &WeChat{
		bot:    bot,
		Wechat: wechat,
	}
}

func (w *WeChat) chatRoomMember(room_name string) (map[string]int, error) {

	stats := make(map[string]int)

	RoomContactList, err := w.Wechat.MembersOfGroup(room_name)
	if err != nil {
		return nil, err
	}

	man := 0
	woman := 0
	none := 0
	for _, v := range RoomContactList {

		member, err := w.Wechat.ContactByGGID(v.GGID)

		if err != nil {
			CLog("[ERRO] 抓取组群用户 [%s] 信息失败: %s... ", v.NickName)
		} else {
			if member.Sex == 1 {
				man++
			} else if member.Sex == 2 {
				woman++
			} else {
				none++
			}
		}

	}

	stats = map[string]int{
		"woman": woman,
		"man":   man,
		"none":  none,
	}

	return stats, nil
}

var x *xiaoice

// 初始化
func (w *WeChat) Construct() error {

	x = newXiaoice(w)

	w.Wechat.Handle(`/login`, func(wechat.Event) {
		if cs, err := w.Wechat.ContactsByNickName(`小冰`); err == nil {
			for _, c := range cs {
				if c.Type == wechat.Offical {
					x.un = c.UserName // 更新小冰的UserName
					break
				}
			}
		}
	})

	return nil
}

// 解析
func (w *WeChat) Process() error {

	//
	w.Wechat.Handle(`/msg`, func(evt wechat.Event) {
		msg := evt.Data.(wechat.EventMsgData)

		if msg.IsGroupMsg {

			if msg.AtMe {
				realcontent := strings.TrimSpace(strings.Replace(msg.Content, "@"+w.Wechat.MySelf.NickName, "", 1))
				if realcontent == "统计人数" {
					stat, err := w.chatRoomMember(msg.FromUserName)
					if err == nil {
						ans := fmt.Sprintf("据统计群里男生 %d 人， 女生 %d 人，未知性别者 %d 人 (ó-ò) ", stat["man"], stat["woman"], stat["none"])

						w.Wechat.SendTextMsg(ans, msg.FromUserName)
					} else {
						w.Wechat.SendTextMsg(err.Error(), msg.FromUserName)
					}
				} else {
					amsg := axiom.Message{
						ToUserName: msg.FromUserName,
						ToAXID: msg.FromGGID,
						Text: realcontent,
					}

					w.bot.ReceiveMessage(amsg)
					x.autoReplay(msg)
				}
			}

		} else {

			amsg := axiom.Message{
				ToUserName: msg.FromUserName,
				ToAXID: msg.FromGGID,
				Text: msg.Content,
			}
			//x.autoReplay(msg)

			w.bot.ReceiveMessage(amsg)
		}

	})

	w.Wechat.Go()
	return nil
}

// 回应
func (w *WeChat) Reply(msg axiom.Message, message string) error {

	w.Wechat.SendTextMsg(message, msg.ToUserName)

	return nil
}
