package main

import (
	"net/http"
	"time"
	"fmt"
	"io/ioutil"
	"encoding/json"
	"bytes"
	"regexp"
	"os"
	"io"
	"os/exec"
	"runtime"
	"encoding/xml"
	"strconv"
	"strings"
	"math"
	"math/rand"
	"net/http/cookiejar"
	"axiom"
)

const (
	appid string = "wx782c26e4c19acffb"

	DefaultUserAgent string = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_11_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/56.0.2924.87 Safari/537.36"
)

type WeChat struct {
	bot *axiom.Robot
	wx *weixin
}

type weixin struct {
	qr_code_path string
	// 本次微信登录需要的UUID
	uuid string
	base_uri string
	redirect_uri string
	uin string
	sid string
	skey string
	pass_ticket string
	device_id string
	synckey string
	SyncKey syncKey
	User User
	BaseRequest  BaseRequest
	syncHost     string

	Users           []string
	InitContactList []User   //谈话的人
	MemberList      []Member //
	ContactList     []Member //好友
	GroupList       []string //群
	GroupMemberList []Member //群友
	PublicUserList  []Member //公众号
	SpecialUserList []Member //特殊账号

	client  *http.Client
}

type User struct {
	UserName          string `json:"UserName"`
	Uin               int64  `json:"Uin"`
	NickName          string `json:"NickName"`
	HeadImgUrl        string `json:"HeadImgUrl" xml:""`
	RemarkName        string `json:"RemarkName" xml:""`
	PYInitial         string `json:"PYInitial" xml:""`
	PYQuanPin         string `json:"PYQuanPin" xml:""`
	RemarkPYInitial   string `json:"RemarkPYInitial" xml:""`
	RemarkPYQuanPin   string `json:"RemarkPYQuanPin" xml:""`
	HideInputBarFlag  int    `json:"HideInputBarFlag" xml:""`
	StarFriend        int    `json:"StarFriend" xml:""`
	Sex               int    `json:"Sex" xml:""`
	Signature         string `json:"Signature" xml:""`
	AppAccountFlag    int    `json:"AppAccountFlag" xml:""`
	VerifyFlag        int    `json:"VerifyFlag" xml:""`
	ContactFlag       int    `json:"ContactFlag" xml:""`
	WebWxPluginSwitch int    `json:"WebWxPluginSwitch" xml:""`
	HeadImgFlag       int    `json:"HeadImgFlag" xml:""`
	SnsFlag           int    `json:"SnsFlag" xml:""`
}

type Member struct {
	Uin              int64
	UserName         string
	NickName         string
	HeadImgUrl       string
	ContactFlag      int
	MemberCount      int
	MemberList       []User
	RemarkName       string
	HideInputBarFlag int
	Sex              int
	Signature        string
	VerifyFlag       int
	OwnerUin         int
	PYInitial        string
	PYQuanPin        string
	RemarkPYInitial  string
	RemarkPYQuanPin  string
	StarFriend       int
	AppAccountFlag   int
	Statues          int
	AttrStatus       int
	Province         string
	City             string
	Alias            string
	SnsFlag          int
	UniFriend        int
	DisplayName      string
	ChatRoomId       int
	KeyWord          string
	EncryChatRoomId  string
}

type syncKey struct {
	Count int      `json:"Count"`
	List  []keyVal `json:"List"`
}

type keyVal struct {
	Key int `json:"Key"`
	Val int `json:"Val"`
}

type BaseRequest struct {
	Uin int
	Sid string
	Skey string
	DeviceID string
}

type BaseResponse struct {
	Ret    int
	ErrMsg string
}

type initResp struct {
	BaseResponse
	User    User
	Skey    string
	SyncKey syncKey
}

func NewWeChat(bot *axiom.Robot, qr_code_path string) *WeChat {
	wx := &weixin{
		qr_code_path: qr_code_path,
	}
	return &WeChat{
		bot: bot,
		wx: wx,
	}
}

// 第一步获取UUID
func (w *WeChat) getUUID(args ...interface{}) bool {
	url := fmt.Sprintf("https://login.weixin.qq.com/jslogin?appid=%s&fun=%s&lang=%s&_=%d", appid, "new", "zh_CN", time.Now().Unix())

	params := make(map[string]interface{})

	data, err := w.httpPost(url, params)
	if err != nil {
		return false
	}

	res := string(data)

	reg := regexp.MustCompile(`window.QRLogin.code = (\d+); window.QRLogin.uuid = "([\S]+)"`)
	req := reg.FindStringSubmatch(res)

	code := req[1]
	if code == "200" {
		w.wx.uuid = req[2]
		return true
	}
	return false
}

// 第二步获取二维码
func (w *WeChat) getQRcode(args ...interface{}) bool {
	url := "https://login.weixin.qq.com/qrcode/" + w.wx.uuid
	params := map[string]interface{} {
		"t": "webwx",
		"_": time.Now().Unix(),
	}

	path := w.wx.qr_code_path + "/qrcode.jpg"
	out, err := os.Create(path)

	resp, err := w.httpPost(url, params)
	if err != nil {
		return false
	}
	_, err = io.Copy(out, bytes.NewReader(resp))
	if err != nil {
		return false
	} else {
		if runtime.GOOS == "darwin" {
			exec.Command("open", path).Run()
		} else {
			go func() {
				fmt.Println("please open on web broswer ip:99250/qrcode")
				http.HandleFunc("/qrcode", func(w http.ResponseWriter, req *http.Request) {
					http.ServeFile(w, req, "qrcode.jpg")
					return
				})
				http.ListenAndServe(":99250", nil)
			}()
		}
		return true
	}
}

// 第三步， 等待登录
func (w *WeChat) waitForLogin(tip int) bool {
	time.Sleep(time.Duration(tip) * time.Second)

	url := fmt.Sprintf("https://login.weixin.qq.com/cgi-bin/mmwebwx-bin/login?tip=%d&uuid=%s&_=%d", tip, w.wx.uuid, time.Now().Unix())

	data, _ := w.httpGet(url)

	reg := regexp.MustCompile(`window.code=(\d+);`)
	res := string(data)

	req := reg.FindStringSubmatch(res)

	if len(req) > 1 {
		code := req[1]
		if code == "201" {
			return true

		} else if code == "200" {
			u_reg := regexp.MustCompile(`window.redirect_uri="(\S+?)";`)
			u_req := u_reg.FindStringSubmatch(res)

			if len(u_req) > 1 {
				r_uri := u_req[1] + "&fun=new"
				w.wx.redirect_uri = r_uri

				bu_reg := regexp.MustCompile(`/`)
				bu_req := bu_reg.FindAllStringIndex(r_uri, -1)

				w.wx.base_uri = r_uri[:bu_req[len(bu_req)-1][0]]

				return true
			}
			return false
		} else if code == "408" {
			CLog(" @@ 登陆超时 @@ ...")
		} else {
			CLog(" @@ 登陆异常 @@ ...")
		}
	}

	return false
}

// 第四步，登录获取Cookie（获取xml中的 skey, wxsid, wxuin, pass_ticket）
func (w *WeChat) login(args ...interface{}) bool {

	data, _ := w.httpGet(w.wx.redirect_uri)

	type result struct {
		Skey        string `xml:"skey"`
		Wxsid       string `xml:"wxsid"`
		Wxuin       string `xml:"wxuin"`
		Pass_ticket string `xml:"pass_ticket"`
	}
	v := result{}
	err := xml.Unmarshal([]byte(data), &v)
	if err != nil {
		fmt.Printf("error: %v", err)
		return false
	}
	w.wx.skey = v.Skey
	w.wx.sid = v.Wxsid
	w.wx.uin = v.Wxuin
	w.wx.pass_ticket = v.Pass_ticket

	w.wx.BaseRequest.Uin, _ = strconv.Atoi(v.Wxuin)
	w.wx.BaseRequest.Sid = v.Wxsid
	w.wx.BaseRequest.Skey = v.Skey
	w.wx.BaseRequest.DeviceID = w.wx.device_id
	return true
}

// 第五步，初始化微信（获取 SyncKey, User 后面的消息监听用）
func (w *WeChat) webWXInit(args ...interface{}) bool {
		url := fmt.Sprintf("%s/webwxinit?pass_ticket=%s&skey=%s&r=%s", w.wx.base_uri, w.wx.pass_ticket, w.wx.skey, time.Now().Unix())
	params := map[string]interface{} {
		"BaseRequest":w.wx.BaseRequest,
	}

	res, err := w.httpPost(url, params)
	if err != nil {
		CLog("[ ERRO ] 初始化微信，访问链接失败：%s ...", err)
		return false
	}
	//log
	ioutil.WriteFile("tmp.txt", []byte(res), 777)
	//log

	var data initResp
	err = json.Unmarshal(res, &data)
	if err != nil {
		CLog("[ ERRO ] 初始化微信，解析失败：%s ...", err)
		return false
	}

	w.wx.User = data.User
	w.wx.SyncKey = data.SyncKey

	w._setsynckey()

	retCode := data.BaseResponse.Ret
	return retCode == 0
}

func (w *WeChat) _setsynckey() {
	keys := []string{}
	for _, keyVal := range w.wx.SyncKey.List {
		keys = append(keys, strconv.Itoa(keyVal.Key)+"_"+strconv.Itoa(keyVal.Val))
	}
	w.wx.synckey = strings.Join(keys, "|")
}

// 第六步，开启微信状态通知
func (w *WeChat) wxStatusNotify(args ...interface{}) bool {
	url := fmt.Sprintf("%s/webwxstatusnotify?lang=zh_CN&pass_ticket=%s", w.wx.base_uri, w.wx.pass_ticket)
	params := map[string]interface{}{
		"BaseRequest": w.wx.BaseRequest,
		"Code": 3,
		"FromUserName": w.wx.User.UserName,
		"ToUserName": w.wx.User.UserName,
		"ClientMsgId": time.Now().Unix(),
	}

	res, err := w.httpPost(url, params)
	if err != nil {
		CLog(" [ ERRO ] 通知状态链接访问失败... ")
		return false
	}

	var data BaseResponse
	err = json.Unmarshal(res, &data)
	if err != nil {
		CLog(" [ ERRO ] 初始化微信，解析失败：%s", err)
		return false
	}

	retCode := data.Ret
	return retCode == 0
}

// 请求群组列表
func (w *WeChat) webGetChatRoomMember(chatroomId string) (map[string]string, error) {
	url := fmt.Sprintf("%s/webwxbatchgetcontact?pass_ticket=%s&skey=%s&r=%s", w.wx.base_uri, w.wx.pass_ticket, w.wx.skey, time.Now().Unix())

	params := make(map[string]interface{})
	params["BaseRequest"] = w.wx.BaseRequest
	params["Count"] = 1
	params["List"] = []map[string]string{}
	l := []map[string]string{}
	params["List"] = append(l, map[string]string{
		"UserName":   chatroomId,
		"ChatRoomId": "",
	})

	members := []string{}
	stats := make(map[string]string)
	res, err := w.httpPost(url, params)
	if err != nil {
		return stats, err
	}

	var data = make(map[string]interface{})
	err = json.Unmarshal(res, data)
	if err != nil {
		//panic("初始化微信，解析失败：" + err)
		return stats, err
	}

	RoomContactList := data["ContactList"].([]interface{})[0].(map[string]interface{})["MemberList"]
	man := 0
	woman := 0
	for _, v := range RoomContactList.([]interface{}) {
		if m, ok := v.([]interface{}); ok {
			for _, s := range m {
				members = append(members, s.(map[string]interface{})["UserName"].(string))
			}
		} else {
			members = append(members, v.(map[string]interface{})["UserName"].(string))
		}
	}
	url = fmt.Sprintf("%s/webwxbatchgetcontact?pass_ticket=%s&skey=%s&r=%s", w.wx.base_uri, w.wx.pass_ticket, w.wx.skey, time.Now().Unix())
	length := 50

	mnum := len(members)
	block := int(math.Ceil(float64(mnum) / float64(length)))
	k := 0
	for k < block {
		offset := k * length
		var l int
		if offset+length > mnum {
			l = mnum
		} else {
			l = offset + length
		}
		blockmembers := members[offset:l]
		params := make(map[string]interface{})
		params["BaseRequest"] = w.wx.BaseRequest
		params["Count"] = len(blockmembers)
		blockmemberslist := []map[string]string{}
		for _, g := range blockmembers {
			blockmemberslist = append(blockmemberslist, map[string]string{
				"UserName":        g,
				"EncryChatRoomId": chatroomId,
			})
		}
		params["List"] = blockmemberslist

		dic, err := w.httpPost(url, params)
		if err == nil {
			userlist := make(map[string]interface{})
			err = json.Unmarshal(dic, userlist)
			if err == nil {
				for _, u := range userlist["ContactList"].([]interface{}) {
					if u.(map[string]interface{})["Sex"].(int) == 1 {
						man++
					} else if u.(map[string]interface{})["Sex"].(int) == 2 {
						woman++
					}
				}
			}
		}
		k++
	}
	stats = map[string]string{
		"woman": strconv.Itoa(woman),
		"man":   strconv.Itoa(man),
	}
	return stats, nil
}

// 消息检查
func (w *WeChat) syncCheck(args ...interface{}) (string, string) {
	urlstr := fmt.Sprintf("https://%s/cgi-bin/mmwebwx-bin/synccheck?r=%d&sid=%s&uin=%s&skey=%s&deviceid=%s&synckey=%s&_=%d", w.wx.syncHost, time.Now().Unix(), w.wx.sid, w.wx.uin, w.wx.skey, w.wx.device_id, w.wx.synckey, time.Now().Unix())

	data, _ := w.httpGet(urlstr)

	reg := regexp.MustCompile(`window.synccheck={retcode:"(\d+)",selector:"(\d+)"}`)

	res := string(data)
	println(res)

	find := reg.FindStringSubmatch(res)
	if len(find) > 2 {
		retcode := find[1]
		selector := find[2]
		return retcode, selector
	} else {
		return "9999", "0"
	}
}

// 同步线路测试
func (w *WeChat) testsynccheck(args ...interface{}) bool {
	hosts := [...]string{
		`webpush.wx.qq.com`,
		`wx2.qq.com`,
		`webpush.wx2.qq.com`,
		`wx8.qq.com`,
		`webpush.wx8.qq.com`,
		`qq.com`,
		`webpush.wx.qq.com`,
		`web2.wechat.com`,
		`webpush.web2.wechat.com`,
		`wechat.com`,
		`webpush.web.wechat.com`,
		`webpush.weixin.qq.com`,
		`webpush.wechat.com`,
		`webpush1.wechat.com`,
		`webpush2.wechat.com`,
		`webpush2.wx.qq.com`}

	for _, host := range hosts {
		CLog("< * > 尝试连接: %s ... ... ", host)
		w.wx.syncHost = host
		code, _ := w.syncCheck()
		if code == `0` {
			return true
		}
		CLog(" [ * ] %s 连接失败...", host)
	}

	return false
}

// 获取新消息
func (w *WeChat) webwxsync(args ...interface{}) (interface{}, error) {
	url := fmt.Sprintf("%s/webwxsync?sid=%s&skey=%s&pass_ticket=%s", w.wx.base_uri, w.wx.sid, w.wx.skey, w.wx.pass_ticket)
	params := make(map[string]interface{})
	params["BaseRequest"] = w.wx.BaseRequest
	params["SyncKey"] = w.wx.SyncKey
	params["rr"] = ^int(time.Now().Unix())
	res, err := w.httpPost(url, params)
	if err != nil{
		return nil, err
	}

	var data = make(map[string]interface{})
	err = json.Unmarshal(res, data)
	if err != nil {
		//panic("初始化微信，解析失败：" + err)
		return nil, err
	}

	retCode := data["BaseResponse"].(map[string]interface{})["Ret"].(int)
	if retCode == 0 {
		w.wx.SyncKey = data["SyncKey"].(syncKey)
		w._setsynckey()
	}
	return data, nil
}

// 发送消息
func (w *WeChat) webWXsendMsg(message string, toUseNname string) bool {
	url := fmt.Sprintf("%s/webwxsendmsg?pass_ticket=%s", w.wx.base_uri, w.wx.pass_ticket)

	clientMsgId := strconv.FormatInt(time.Now().Unix(), 10) + strconv.Itoa(rand.Int())[3:7]

	params := make(map[string]interface{})
	params["BaseRequest"] = w.wx.BaseRequest

	msg := make(map[string]interface{})
	msg["Type"] = 1
	msg["Content"] = message
	msg["FromUserName"] = w.wx.User.UserName
	msg["ToUserName"] = toUseNname
	msg["LocalID"] = clientMsgId
	msg["ClientMsgId"] = clientMsgId
	params["Msg"] = msg

	_, err := w.httpPost(url, params)
	if err != nil {
		return false
	} else {
		return true
	}
}

// 处理消息
func (w *WeChat) handleMsg(r interface{}) {
	myNickName := w.wx.User.NickName
	for _, msg := range r.(map[string]interface{})["AddMsgList"].([]interface{}) {
		// fmt.Println("[*] 你有新的消息，请注意查收")
		// msg = msg.(map[string]interface{})
		msgType := msg.(map[string]interface{})["MsgType"].(int)
		fromUserName := msg.(map[string]interface{})["FromUserName"].(string)
		// name = self.getUserRemarkName(msg['FromUserName'])
		content := msg.(map[string]interface{})["Content"].(string)
		content = strings.Replace(content, "&lt;", "<", -1)
		content = strings.Replace(content, "&gt;", ">", -1)
		content = strings.Replace(content, " ", " ", 1)
		// msgid := msg.(map[string]interface{})["MsgId"].(int)
		if msgType == 1 {
			var ans string
			var err error
			if fromUserName[:2] == "@@" {
				CLog(" # * # 收到群消息：" + content + "|0045")
				contentSlice := strings.Split(content, ":<br/>")
				// people := contentSlice[0]
				content = contentSlice[1]
				if strings.Contains(content, "@"+myNickName) {
					realcontent := strings.TrimSpace(strings.Replace(content, "@"+myNickName, "", 1))
					CLog(" # * # 收到群消息：" + realcontent + "|0046")
					if realcontent == "统计人数" {
						stat, err := w.webGetChatRoomMember(fromUserName)
						if err == nil {
							ans = " # * # 据统计群里男生" + stat["man"] + "人，女生" + stat["woman"] + "人 (ó㉨ò)"
						}
					}
				}
			}

			if err != nil {
				CLog("[ ERRO ] : " + err.Error())
			} else if ans != "" {
				go w.webWXsendMsg(ans, fromUserName)
			}
		} else if msgType == 51 {
			CLog(" # * # 成功截获微信初始化消息")
		}
	}
}

// get 方法
func (w *WeChat) httpGet(url string) ([]byte, error) {

	request, _ := http.NewRequest("GET", url, nil)

	request.Header.Add("Referer", "https://wx.qq.com/")

	request.Header.Add("User-agent", DefaultUserAgent)

	resp, err := w.wx.client.Do(request)

	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return ioutil.ReadAll(resp.Body)
}

// post 方法
func (w *WeChat) httpPost(url string, params map[string]interface{}) ([]byte, error) {
	postJson, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}

	postData := bytes.NewBuffer(postJson)

	request, err := http.NewRequest("POST", url, postData)

	if err != nil {
		return nil, err
	}

	request.Header.Set("Content-Type", "application/json;charset=utf-8")

	request.Header.Add("Referer", "https://wx.qq.com/")

	request.Header.Add("User-agent", DefaultUserAgent)

	resp, err := w.wx.client.Do(request)

	if err != nil || resp == nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func (w *WeChat) _run(desc string, f func(...interface{}) bool, args ...interface{}) {
	start := time.Now().UnixNano()
	CLog(desc)
	var result bool
	if len(args) > 1 {
		result = f(args)
	} else if len(args) == 1 {
		result = f(args[0])
	} else {
		result = f()
	}
	useTime := fmt.Sprintf("%.5f", (float64(time.Now().UnixNano()-start) / 1000000000))
	if result {
		CLog(" @@ 成功 @@ , 用时 < " + useTime + " > 秒")
	} else {
		CLog(" ( 失败 ) # 退出程序 # ...")
		os.Exit(1)
	}
}

func (w *WeChat) _init() {
	gCookieJar, _ := cookiejar.New(nil)
	httpclient := http.Client{
		CheckRedirect: nil,
		Jar:           gCookieJar,
	}
	w.wx.client = &httpclient
	rand.Seed(time.Now().Unix())
	str := strconv.Itoa(rand.Int())
	w.wx.device_id = "e" + str[2:17]
}

// 初始化
func (w *WeChat) Construct() error {
	CLog(" # ** #  微信网页版... 开动 ")
	w._init()
	w._run(" # ** # 获取 uuid ... ", w.getUUID)
	w._run(" # ** # 正在获取 二维码 ... ", w.getQRcode)
	CLog(" # ** # 请使用微信扫描二维码以登录 ... ")
	for {
		if w.waitForLogin(1) == false {
			continue
		}
		CLog(" # ** # 请在手机上点击确认以登录 ... ")
		if w.waitForLogin(0) == false {
			continue
		}
		break
	}
	w._run(" # ** # 正在登录 ... ", w.login)
	w._run(" # ** # 微信初始化 ... ", w.webWXInit)
	w._run(" # ** # 开启状态通知 ... ", w.wxStatusNotify)
	w._run(" # ** # 进行同步线路测试 ... ", w.testsynccheck)

	return nil
}

// 解析
func (w *WeChat) Process() error {
	return nil
}

// 回应
func (w *WeChat) Reply(msg axiom.Message, message string) error {
	for {
		retcode, selector := w.syncCheck()
		if retcode == "1100" {
			CLog(" # ** # 你在手机上登出了微信，债见")
			break
		} else if retcode == "1101" {
			CLog(" # ** # 你在其他地方登录了 WEB 版微信，债见")
			break
		} else if retcode == "0" {
			if selector == "2" {
				r, _ := w.webwxsync()
				switch r.(type) {
				case bool:
				default:
					w.handleMsg(r)
				}
			} else if selector == "0" {
				time.Sleep(1)
			} else if selector == "6" || selector == "4" {
				w.webwxsync()
				time.Sleep(1)
			}
		}
	}
	return nil
}

