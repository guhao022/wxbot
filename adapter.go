package main

import (
	"axiom"
	"fmt"
	"strings"
	"wxbot/wechat"
)

type WeChat struct {
	bot    *axiom.Robot
	wechat *wechat.WeChat
}

func NewWeChat(bot *axiom.Robot) *WeChat {

	wechat, err := wechat.AwakenNewBot(nil)
	if err != nil {
		panic(err)
	}

	return &WeChat{
		bot:    bot,
		wechat: wechat,
	}
}

func (w *WeChat) chatRoomMember(room_name string) (map[string]int, error) {

	stats := make(map[string]int)

	RoomContactList, err := w.wechat.MembersOfGroup(room_name)
	if err != nil {
		return nil, err
	}

	man := 0
	woman := 0
	none := 0
	for _, v := range RoomContactList {

		member, err := w.wechat.ContactByGGID(v.GGID)

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

// 初始化
func (w *WeChat) Construct() error {

	w.wechat.Handle(`/login`, func(wechat.Event) {})

	return nil
}

// 解析
func (w *WeChat) Process() error {

	//
	w.wechat.Handle(`/msg`, func(evt wechat.Event) {
		msg := evt.Data.(wechat.EventMsgData)

		if msg.IsGroupMsg {
			if msg.AtMe {
				realcontent := strings.TrimSpace(strings.Replace(msg.Content, "@"+w.wechat.MySelf.NickName, "", 1))
				if realcontent == "统计人数" {
					stat, err := w.chatRoomMember(msg.FromUserName)
					if err == nil {
						ans := fmt.Sprintf("据统计群里男生 %d 人， 女生 %d 人，未知性别者 %d 人 (ó-ò) ", stat["man"], stat["woman"], stat["none"])

						w.wechat.SendTextMsg(ans, msg.FromUserName)
					} else {
						w.wechat.SendTextMsg(err.Error(), msg.FromUserName)
					}
				} else {
					amsg := axiom.Message{
						User: msg.FromUserName,
						Text: realcontent,
					}

					w.bot.ReceiveMessage(amsg)
				}
			}

		} else {
			amsg := axiom.Message{
				User: msg.FromUserName,
				Text: msg.Content,
			}

			w.bot.ReceiveMessage(amsg)
		}

	})

	w.wechat.Go()
	return nil
}

// 回应
func (w *WeChat) Reply(msg axiom.Message, message string) error {

	w.wechat.SendTextMsg(message, msg.User)

	return nil
}
