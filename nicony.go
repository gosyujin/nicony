package main

import (
	"encoding/json"
	"encoding/xml"
	"flag"
	"fmt"
	"github.com/PuerkitoBio/goquery"
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

// オプション情報
type Option struct {
	IsAnsi        *bool // ログ出力をAnsiカラーにするか
	IsProgressBar *bool // ダウンロード時プログレスバーを表示するか
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
	Tags          []Tag  `xml:"tags"`           //18 タグ //TODO 同じタグが複数ある時の取得方法
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

func initLogger(o Option) {
	// seelog設定
	var logConfig string
	if *o.IsAnsi {
		logConfig = `
		  <seelog type="adaptive" mininterval="200000000" maxinterval="1000000000" critmsgcount="5">
		    <formats>
		      <format id="console" format="%EscM(36)[nicony]%EscM(39) %EscM(32)%Date(2006-01-02T15:04:05Z07:00)%EscM(39) %EscM(33)[%File:%FuncShort:%Line]%EscM(39) %EscM(46)[%LEVEL]%EscM(49) %Msg%n" />
		      <format id="plane" format="[nicony] %Date(2006-01-02T15:04:05Z07:00) [%File:%FuncShort:%Line] [%LEVEL] %Msg%n" />
		    </formats>
		    <outputs>
		      <filter formatid="console" levels="debug,info,warn,error,critical">
		        <console />
		      </filter>
		      <filter formatid="plane" levels="trace,debug,info,warn,error,critical">
		        <rollingfile filename="./log/log.txt" type="size" maxsize="1024000" maxrolls="500" />
		      </filter>
		    </outputs>
		  </seelog>`
	} else {
		logConfig = `
		  <seelog type="adaptive" mininterval="200000000" maxinterval="1000000000" critmsgcount="5">
		    <formats>
		      <format id="plane" format="[nicony] %Date(2006-01-02T15:04:05Z07:00) [%File:%FuncShort:%Line] [%LEVEL] %Msg%n" />
		    </formats>
		    <outputs>
		      <filter formatid="plane" levels="debug,info,warn,error,critical">
		        <console />
		      </filter>
		      <filter formatid="plane" levels="trace,debug,info,warn,error,critical">
		        <rollingfile filename="./log/log.txt" type="size" maxsize="1024000" maxrolls="500" />
		      </filter>
		    </outputs>
		  </seelog>`
	}

	logger, _ := log.LoggerFromConfigAsBytes([]byte(logConfig))
	log.ReplaceLogger(logger)
}

func main() {
	o := Option{}
	o.IsAnsi = flag.Bool("ansi", true, "Output Ansi color")
	o.IsProgressBar = flag.Bool("pb", true, "Show progress bar")
	flag.Parse()

	initLogger(o)
	defer log.Flush()
	log.Info("nicony ver.0.3")

	login()

	// ニコレポから動画リスト取得
	var links []string
	links = getNicorepo(getNicorepoUrl, links)

	for _, url := range links {
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
			continue
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
				continue
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

func getNicorepo(getNicorepoUrl string, links []string) []string {
	client := http.Client{Jar: jar}
	log.Trace("getNicorepo URL: " + getNicorepoUrl)
	res, _ := client.Get(getNicorepoUrl)
	body, _ := ioutil.ReadAll(res.Body)
	defer res.Body.Close()
	doc, _ := goquery.NewDocumentFromReader(strings.NewReader(string(body)))

	// タイムライン
	nextPageLink := ""
	doc.Find(".nicorepo-page").Each(func(_ int, s *goquery.Selection) {
		// マイリスト
		s.Find(".log-user-mylist-add .log-target-info").Each(func(_ int, s *goquery.Selection) {
			log.Trace(s.Find("a").Text())
			link, _ := s.Find("a").Attr("href")
			buf := strings.Split(link, "/")
			number := buf[len(buf)-1]
			links = append(links, number)
		})
		// 公式が動画投稿
		s.Find(".log-community-video-upload .log-target-info").Each(func(_ int, s *goquery.Selection) {
			log.Trace(s.Find("a").Text())
			link, _ := s.Find("a").Attr("href")
			buf := strings.Split(link, "/")
			number := buf[len(buf)-1]
			links = append(links, number)
		})
		// ユーザーが動画投稿
		s.Find(".log-user-video-upload .log-target-info").Each(func(_ int, s *goquery.Selection) {
			log.Trace(s.Find("a").Text())
			link, _ := s.Find("a").Attr("href")
			buf := strings.Split(link, "/")
			number := buf[len(buf)-1]
			links = append(links, number)
		})
		// xx再生を達成
		s.Find(".log-user-video-round-number-of-view-counter .log-target-info").Each(func(_ int, s *goquery.Selection) {
			log.Trace(s.Find("a").Text())
			link, _ := s.Find("a").Attr("href")
			buf := strings.Split(link, "/")
			number := buf[len(buf)-1]
			links = append(links, number)
		})
		// ニコレポ最後
		s.Find(".no-next-page").Each(func(_ int, s *goquery.Selection) {
			log.Trace(".no-next")
		})

		// 過去のニコレポ再帰呼び出し
		s.Find(".next-page-link").Each(func(_ int, s *goquery.Selection) {
			log.Trace(".next")
			link, _ := s.Attr("href")
			nextPageLink = link
		})
	})
	if nextPageLink == "" {
		return links
	} else {
		return getNicorepo(nicovideojpUrl+nextPageLink, links)
	}
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
	log.Debug(res.Status)

	body, _ := ioutil.ReadAll(res.Body)
	defer res.Body.Close()

	return body
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

func downloadVideo(filepath string, videoUrl string, nicovideo NicovideoThumbResponse, o Option) {
	client := http.Client{Jar: jar}
	watchUrl := nicovideo.Thumb.WatchUrl
	size, _ := strconv.Atoi(nicovideo.Thumb.SizeHigh)

	// videoUrlにアクセスする前にいったんwatchUrlをgetする必要がある http://n-yagi.0r2.net/script/2009/12/nico2downloader.html
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
