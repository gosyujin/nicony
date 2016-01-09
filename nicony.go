package main

import (
	"encoding/json"
	"encoding/xml"
	"flag"
	"fmt"
	"github.com/cheggaaa/pb"
	log "github.com/cihub/seelog"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	loginUrl        = "https://secure.nicovideo.jp/secure/login"
	nicovideojpUrl  = "http://www.nicovideo.jp/"
	getNicorepoUrl  = "http://www.nicovideo.jp/my/top/all?innerPage=1&mode=next_page"
	getThumbinfoUrl = "http://ext.nicovideo.jp/api/getthumbinfo/"
	getFlvUrl       = "http://flapi.nicovideo.jp/api/getflv/"
	getThreadKeyUrl = "http://flapi.nicovideo.jp/api/getthreadkey/"
)

// ログイン情報
type Account struct {
	Mail     string
	Password string
}

// 指定された動画のFLV保管URLの情報 http://dic.nicovideo.jp/a/ニコニコ動画api
type FlvInfo struct {
	ThreadId         string //1 コメントDLで使う
	L                string //2 コメントDLで使う、60で割って+1して使う
	Url              string //3 動画DLで使う
	Ms               string //4 コメントDLで使う
	MsSub            string //5
	UserId           string //6 コメントDLで使う
	IsPremium        string //7 (プレミアムなら1)
	Nickname         string //8
	Time             string //9
	Done             string //10
	NgRv             string //11
	Hms              string //12
	Hmsp             string //13
	Hmst             string //14
	Hmstk            string //15
	UserKey          string //16
	NeedsKey         string //17 公式放送のみ存在？
	OptionalThreadId string //18 公式放送のみ存在？
	NgCh             string //19 公式放送のみ存在？
}

// 指定された動画の動画情報 http://dic.nicovideo.jp/a/ニコニコ動画api
type NicovideoThumbResponse struct {
	Thumb Thumb `xml:"thumb"`
}

type Thumb struct {
	VideoId       string `xml:"video_id"`       //1 動画ID
	Title         string `xml:"title"`          //2 動画タイトル
	Description   string `xml:"description"`    //3 動画説明文
	ThumbnailUrl  string `xml:"thumbnail_url"`  //4 サムネイルURL
	FirstRetrieve string `xml:"first_retrieve"` //5 投稿日時
	Length        string `xml:"length"`         //6 動画の再生時間
	MovieType     string `xml:"movie_type"`     //7 動画の形式。FlashVideo形式ならflv、MPEG-4形式ならmp4、ニコニコムービーメーカーはswf
	SizeHigh      string `xml:"size_high"`      //8 動画サイズ
	SizeLow       string `xml:"size_low"`       //9 低画質モード時の動画サイズ
	ViewCounter   string `xml:"view_counter"`   //10 再生数
	CommentNum    string `xml:"comment_num"`    //11 コメント数
	MylistCounter string `xml:"mylist_counter"` //12 マイリスト数
	LastResBody   string `xml:"last_res_body"`  //13 ブログパーツなどに表示される最新コメント
	WatchUrl      string `xml:"watch_url"`      //14 視聴URL
	ThumbType     string `xml:"thumb_type"`     //15 動画ならvideo、マイメモリーならmymemory
	Embeddable    string `xml:"embeddable"`     //16 外部プレイヤーで再生禁止(1)か可能(0)
	NoLivePlay    string `xml:"no_live_play"`   //17 ニコニコ生放送で再生禁止(1)か可能(0)
	Tags          []Tag  `xml:"tags"`           //18 タグ
	UserId        string `xml:"user_id"`        //19 ユーザID
	UserNickname  string `xml:"user_nickname"`  //20 ユーザニックネーム
	UserIconUrl   string `xml:"user_icon_url"`  //21 ユーザアイコン
	ChId          string `xml:"ch_id"`          //22 チャンネルID
	ChName        string `xml:"ch_name"`        //23 チャンネル名
	ChIconUrl     string `xml:"ch_icon_url"`    //24 チャンネルアイコン
}

type Tag struct {
	Tag []string `xml:"tag"`
}

//コメントDLに使う
type ThreadKeyInfo struct {
	ThreadKey string //1 コメントDLで使う、必ず空？
	Force184  string //2 コメントDLで使う、必ず1？
}

// cookie
var jar, _ = cookiejar.New(nil)

// オプション情報
type Option struct {
	IsAnsi        *bool   // ログ出力をAnsiカラーにするか
	IsProgressBar *bool   // ダウンロード時プログレスバーを表示するか
	IsVersion     *bool   // バージョン表示
	LogLevel      *string // ログレベル
	VideoId       *string // ビデオID
}

func optionParser() Option {
	o := Option{}
	o.IsAnsi = flag.Bool("ansi", true, "Enable Ansi color")
	o.LogLevel = flag.String("l", "debug", "Log level")
	o.VideoId = flag.String("id", "", "Video ID (ex.sm123456789)")
	o.IsProgressBar = flag.Bool("pb", true, "Show progress bar")
	o.IsVersion = flag.Bool("v", false, "Show version")
	flag.Parse()

	return o
}

func main() {
	o := optionParser()

	if *o.IsVersion {
		fmt.Println(getVersion())
		os.Exit(0)
	}

	initLogger(o)
	defer log.Flush()

	log.Info(getVersion())

	login()

	if *o.VideoId == "" {
		// ニコレポページから動画リスト取得
		var links []string
		links = getNicorepo(getNicorepoUrl, links)

		for _, videoId := range links {
			download(videoId, o)
		}
	} else {
		download(*o.VideoId, o)
	}

}

func download(url string, o Option) {
	time.Sleep(time.Millisecond * 15000)
	log.Info("===================================================")
	// flv保管情報取得
	flvInfo := getFlvInfo(getFlvUrl + url)
	log.Tracef("%#v", flvInfo)
	// 動画情報取得(未ログインでも取得できる)
	nicovideo := getThumb(getThumbinfoUrl + url)

	title := strings.Replace(nicovideo.Thumb.Title, "/", "", -1)
	chName := nicovideo.Thumb.ChName
	nickname := nicovideo.Thumb.UserNickname
	movieType := nicovideo.Thumb.MovieType
	sizeHigh := nicovideo.Thumb.SizeHigh

	log.Info("target: " + title)

	if flvInfo.Url == "" {
		log.Warn("flvInfo.Url is EMPTY.無料期間終了か、元から有料っぽい")
		return
	}

	filepath := "dest"
	if chName == "" {
		filepath = filepath + "/user/" + nickname
	} else {
		filepath = filepath + "/channel/" + chName
	}
	dest := filepath + "/" + title

	fi, _ := os.Stat(dest + "." + movieType)
	fullsize, _ := strconv.Atoi(sizeHigh)
	if fi == nil {
		log.Info("new download: " + dest)
	} else {
		if int(fi.Size()) == fullsize {
			log.Warn("video is already EXIST.動画は既に存在している: " + dest)
			return
		} else {
			log.Warn("redownload: " + dest)
			os.Remove(dest + "." + movieType)
		}
	}

	log.Info("make dir: " + filepath)
	os.MkdirAll(filepath, 0711)

	// 動画情報ファイル出力
	buf, _ := xml.MarshalIndent(nicovideo, "", "  ")
	write(dest+".txt", buf)

	// コメント取得
	comment := getComment(flvInfo)
	// コメントファイル出力
	write(dest+".xml", comment)

	// 動画ファイル書き込み
	downloadVideo(dest+"."+nicovideo.Thumb.MovieType, flvInfo.Url, nicovideo, o)
}

func login() {
	log.Info("read ./account.json")
	jsonstring, _ := ioutil.ReadFile("./account.json")
	a := Account{}
	json.Unmarshal(jsonstring, &a)
	mail := a.Mail
	password := a.Password

	log.Debug("login URL: " + loginUrl)
	client := http.Client{Jar: jar}
	res, _ := client.PostForm(
		loginUrl,
		url.Values{"mail": {mail}, "password": {password}},
	)
	log.Debug(res.Status)
}

func getThumb(getThumbinfoUrl string) NicovideoThumbResponse {
	log.Debug("getThumbinfo URL: " + getThumbinfoUrl)

	res, _ := http.Get(getThumbinfoUrl)
	log.Debug(res.Status)
	body, _ := ioutil.ReadAll(res.Body)
	defer res.Body.Close()

	nicovideo := NicovideoThumbResponse{Thumb{}}
	xml.Unmarshal(body, &nicovideo)
	log.Tracef("%#v", nicovideo)

	return nicovideo
}

func getFlvInfo(getFlvUrl string) FlvInfo {
	client := http.Client{Jar: jar}

	log.Debug("get getFlvUrl " + getFlvUrl)
	res, _ := client.Get(getFlvUrl)
	body, _ := ioutil.ReadAll(res.Body)
	log.Debug(res.Status)
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
		case "needs_key":
			f.NeedsKey = value
		case "optional_thread_id":
			f.OptionalThreadId = value
		case "ng_ch":
			f.NgCh = value
		default:
			log.Warn("unknown parameter: " + key + " value is " + value)
			// closedがあり、かつ1だと不正終了っぽい
		}
	}

	//TODO 必須パラメータ存在チェック
	//log.Debug(f.ThreadId)
	//log.Debug(f.Url)
	//log.Debug(f.Ms)
	//log.Debug(f.UserId)

	return f
}

func getComment(flvInfo FlvInfo) []byte {
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
	log.Trace(packetXml)

	client := http.Client{Jar: jar}
	messageServer, _ := url.QueryUnescape(flvInfo.Ms)
	log.Debug("comment server URL: " + messageServer)
	res, _ := client.Post(
		messageServer,
		"application/x-www-form-urlencoded",
		strings.NewReader(packetXml),
	)
	log.Tracef("%#v", res)
	log.Debug(res.Status)

	body, _ := ioutil.ReadAll(res.Body)
	defer res.Body.Close()

	return body
}

func getThreadKeyInfo(threadId string) ThreadKeyInfo {
	client := http.Client{Jar: jar}

	threadKeyServer := getThreadKeyUrl + "?thread=" + threadId
	log.Debug("threadKey server URL: " + threadKeyServer)

	res, _ := client.Get(threadKeyServer)
	body, _ := ioutil.ReadAll(res.Body)
	defer res.Body.Close()

	log.Tracef("%#v", res)
	log.Debug(res.Status)

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

	log.Tracef("%#v", t)
	return t
}

func downloadVideo(filepath string, videoUrl string, nicovideo NicovideoThumbResponse, o Option) {
	client := http.Client{Jar: jar}
	watchUrl := nicovideo.Thumb.WatchUrl
	size, _ := strconv.Atoi(nicovideo.Thumb.SizeHigh)

	// videoUrlにアクセスする前にいったんwatchUrlをgetしてCookieを取得している必要がある
	// http://n-yagi.0r2.net/script/2009/12/nico2downloader.html
	// http://sekai.hatenablog.jp/entry/2013/02/26/164500
	client.Get(watchUrl)

	videoUrlDecode, _ := url.QueryUnescape(videoUrl)
	log.Debug("video server URL: " + videoUrlDecode)
	res, _ := client.Get(videoUrlDecode)
	log.Debug(res.Status)
	defer res.Body.Close()

	log.Info("download: " + filepath)
	// File open
	file, _ := os.OpenFile(filepath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
	defer file.Close()

	copy(size, file, res.Body, o)
}

func copy(size int, filepath io.Writer, source io.Reader, o Option) {
	time.Sleep(time.Millisecond * 15000)
	if *o.IsProgressBar {
		// プログレスバー
		progressBar := pb.New(size)
		progressBar.SetUnits(pb.U_BYTES)
		progressBar.SetRefreshRate(time.Millisecond * 10)
		progressBar.ShowCounters = true
		progressBar.ShowTimeLeft = true
		progressBar.ShowSpeed = true
		progressBar.SetMaxWidth(80)
		progressBar.Start()

		writer := io.MultiWriter(filepath, progressBar)
		io.Copy(writer, source)
		progressBar.FinishPrint("download complete")
	} else {
		writer := io.MultiWriter(filepath)
		io.Copy(writer, source)
	}

	time.Sleep(time.Millisecond * 15000)
}

func write(filepath string, body []byte) {
	log.Info("write file: " + filepath)
	ioutil.WriteFile(filepath, body, os.ModePerm)
}
