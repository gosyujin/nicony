package main

import (
	"encoding/json"
	"fmt"
	log "github.com/cihub/seelog"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"strconv"
	"strings"
)

const (
	loginUrl        = "https://secure.nicovideo.jp/secure/login"
	getFlvUrl       = "http://flapi.nicovideo.jp/api/getflv/"
	getThreadKeyUrl = "http://flapi.nicovideo.jp/api/getthreadkey/"
)

type Account struct {
	Mail     string
	Password string
}

type FlvInfo struct {
	ThreadId  string //1 コメントDLで使う、動画DLで使う
	L         string //2 コメントDLで使う、60で割って+1して使う
	Url       string //3 動画DLで使う
	Ms        string //4 コメントDLで使う、動画DLで使う
	MsSub     string //5
	UserId    string //6 コメントDLで使う、動画DLで使う
	IsPremium string //7 (プレミアムなら1)
	Nickname  string //8
	Time      string //9
	Done      string //10
	NgRv      string //11
	Hms       string //12
	Hmsp      string //13
	Hmst      string //14
	Hmstk     string //15
	UserKey   string //16
}

//コメントDLに使う
type ThreadKeyInfo struct {
	ThreadKey string //1 コメントDLで使う、必ず空？
	Force184  string //2 コメントDLで使う、必ず1？
}

// cookie
var jar, _ = cookiejar.New(nil)

// seelog設定
const logConfig = `
<seelog type="adaptive" mininterval="200000000" maxinterval="1000000000" critmsgcount="5">
  <formats>
    <format id="main" format="%EscM(36)[nigo]%EscM(39) %Date(2006-01-02T15:04:05.999999999Z07:00) [%File:%FuncShort:%Line] %EscM(46)[%LEV]%EscM(49) %Msg%n" />
  </formats>
  <outputs formatid="main">
    <filter levels="trace,debug,info,warn,error,critical">
      <console />
    </filter>
    <filter levels="info,warn,error,critical">
      <rollingfile filename="./log.log" type="size" maxsize="102400" maxrolls="1" />
    </filter>
  </outputs>
</seelog>`

func initLogger() {
	logger, err := log.LoggerFromConfigAsBytes([]byte(logConfig))
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	log.ReplaceLogger(logger)
}

func main() {
	initLogger()
	defer log.Flush()

	url := "sm26477759"

	login()
	flvInfo := getFlvInfo(getFlvUrl + url)

	getComment(flvInfo)
}

func login() {
	log.Debug("read ./account.json")
	jsonstring, _ := ioutil.ReadFile("./account.json")
	a := Account{}
	json.Unmarshal(jsonstring, &a)
	mail := a.Mail
	password := a.Password

	log.Debug("login " + loginUrl)
	client := http.Client{Jar: jar}
	res, _ := client.PostForm(
		loginUrl,
		url.Values{"mail": {mail}, "password": {password}},
	)
	log.Debug(res.Header)
}

func getComment(flvInfo FlvInfo) {
	threadKeyInfo := getThreadKeyInfo(flvInfo.ThreadId)

	temp, _ := strconv.Atoi(flvInfo.L)
	minutes := temp/60 + 1
	packetXml := fmt.Sprintf(
		`<packet>
<thread thread="%v" user_id="%v"
  threadkey="%v" force_184="%v"
  scores="1" version="20090904" res_from="-1000"
  with_global="1">
</thread>
<thread_leaves thread="%v#" user_id="%v"
  threadkey="%v" force_184="%v"
  scores="1">
    0-%v:100,1000
</thread_leaves>
</packet>`,
		flvInfo.ThreadId, flvInfo.UserId,
		threadKeyInfo.ThreadKey, threadKeyInfo.Force184,
		flvInfo.ThreadId, flvInfo.UserId,
		threadKeyInfo.ThreadKey, threadKeyInfo.Force184,
		minutes)
	log.Debug(packetXml)

	client := http.Client{Jar: jar}
	messageServer, _ := url.QueryUnescape(flvInfo.Ms)
	log.Debug("post message server: " + messageServer)
	res, _ := client.Post(
		messageServer,
		"application/x-www-form-urlencoded",
		strings.NewReader(packetXml),
	)
	body, _ := ioutil.ReadAll(res.Body)
	defer res.Body.Close()

	log.Debug(res.Header)
	log.Info("write comment file: test.xml")
	ioutil.WriteFile("./test.xml", body, os.ModePerm)
}

func getThreadKeyInfo(threadId string) ThreadKeyInfo {
	client := http.Client{Jar: jar}
	res, _ := client.Get(getThreadKeyUrl + "?thread=" + threadId)
	body, _ := ioutil.ReadAll(res.Body)
	defer res.Body.Close()

	//レスポンスをクエリパラメータ毎に分割
	t := ThreadKeyInfo{}
	for _, param := range strings.Split(string(body), "&") {
		temp := strings.Split(param, "=")
		key := temp[0]
		value := temp[1]

		switch key {
		case "threadkey":
			t.ThreadKey = value
		case "force_184":
			t.Force184 = value
		default:
			log.Warn("unknown parameter: " + key + " value is " + value)
		}
	}
	return t
}

func getFlvInfo(getFlvUrl string) FlvInfo {
	client := http.Client{Jar: jar}
	log.Debug("get getFlvUrl " + getFlvUrl)
	res, _ := client.Get(getFlvUrl)
	body, _ := ioutil.ReadAll(res.Body)
	defer res.Body.Close()

	//レスポンスをクエリパラメータ毎に分割
	f := FlvInfo{}
	for _, param := range strings.Split(string(body), "&") {
		temp := strings.Split(param, "=")
		key := temp[0]
		value := temp[1]

		switch key {
		case "thread_id":
			f.ThreadId = value
		case "l":
			f.L = value
		case "url":
			f.Url = value
		case "ms":
			f.Ms = value
		case "ms_sub":
			f.MsSub = value
		case "user_id":
			f.UserId = value
		case "is_premium":
			f.IsPremium = value
		case "nickname":
			f.Nickname = value
		case "time":
			f.Time = value
		case "done":
			f.Done = value
		case "ng_rv":
			f.NgRv = value
		case "hms":
			f.Hms = value
		case "hmsp":
			f.Hmsp = value
		case "hmst":
			f.Hmst = value
		case "hmstk":
			f.Hmstk = value
		case "userkey":
			f.UserKey = value
		default:
			log.Warn("unknown parameter: " + key + " value is " + value)
		}
	}

	//TODO 必須パラメータ存在チェック
	//log.Debug(f.ThreadId)
	//log.Debug(f.Url)
	//log.Debug(f.Ms)
	//log.Debug(f.UserId)

	return f
}
