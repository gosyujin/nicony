package main

import (
	"encoding/json"
	"encoding/xml"
	"flag"
	"fmt"
	log "github.com/cihub/seelog"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"strings"
)

const (
	loginUrl        = "https://secure.nicovideo.jp/secure/login"
	nicovideojpUrl  = "http://www.nicovideo.jp/"
	getNicorepoUrl  = "http://www.nicovideo.jp/api/nicorepo/timeline/my/all?client_app=pc_myrepo"
	getMylistUrl    = "http://www.nicovideo.jp/mylist/"
	getThumbinfoUrl = "http://ext.nicovideo.jp/api/getthumbinfo/"
	getFlvUrl       = "http://flapi.nicovideo.jp/api/getflv/"
	getThreadKeyUrl = "http://flapi.nicovideo.jp/api/getthreadkey/"
)

// cookie
var jar, _ = cookiejar.New(nil)

// ログイン情報
type Account struct {
	Mail     string
	Password string
}

// オプション情報
type Option struct {
	IsAnsi          *bool   // ログ出力をAnsiカラーにするか
	IsProgressBar   *bool   // ダウンロード時プログレスバーを表示するか
	IsVersion       *bool   // バージョン表示
	IsForce         *bool   // 強制ダウンロード
	LogLevel        *string // ログレベル
	LogDestination  *string // ログ出力場所
	VideoId         *string // ビデオID
	MylistId        *string // マイリストID
	Destination     *string // 出力先
	AccountFilepath *string // ログイン情報ファイルのパス
	Mail            *string // ログイン用メールアドレス
	Password        *string // ログイン用パスワード
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
	log.Info("video Destination path: " + *o.Destination)
	log.Info("log Destination path: " + *o.LogDestination)

	login(o)

	if *o.VideoId != "" {
		// 引数に指定された動画取得
		download(*o.VideoId, o)
	} else if *o.MylistId != "" {
		// 引数に指定されたマイリスト取得
		var links []string
		links = getMylist(getMylistUrl + *o.MylistId + "?rss=2.0")

		for _, videoId := range links {
			download(videoId, o)
		}
	} else {
		// ニコレポページから動画リスト取得
		var links []string
		links = getNicorepo(getNicorepoUrl, "", links)

		for _, videoId := range links {
			download(videoId, o)
		}
	}
}

func optionParser() Option {
	o := Option{}
	o.IsAnsi = flag.Bool("ansi", true, "Enable Ansi color")
	o.IsProgressBar = flag.Bool("pb", true, "Show progress bar")
	o.IsVersion = flag.Bool("v", false, "Show version")
	o.IsForce = flag.Bool("f", false, "Force download")
	o.LogLevel = flag.String("l", "debug", "Log level")
	o.LogDestination = flag.String("logdest", "./var/log/nicony.log", "Log destination path")
	o.VideoId = flag.String("id", "", "Video ID ex.sm123456789")
	o.MylistId = flag.String("mylist", "", "Mylist ID ex.123456789")
	o.Destination = flag.String("d", "./dest", "Destination path")
	o.AccountFilepath = flag.String("a", "./account.json", "Login account setting file")
	o.Mail = flag.String("mail", "", "Login mailaddress (-m MAILADDRESS -p PASSWORD)")
	o.Password = flag.String("password", "", "Login password (-m MAILADDRESS -p PASSWORD)")

	flag.Parse()

	return o
}

func login(o Option) {
	a := Account{}
	if *o.Mail != "" && *o.Password != "" {
		log.Info("read mailaddress and password args")
		a.Mail = *o.Mail
		a.Password = *o.Password
	} else {
		log.Info("read accountFile: " + *o.AccountFilepath)
		jsonstring, err := ioutil.ReadFile(*o.AccountFilepath)
		if err != nil {
			log.Error(err)
			sleep(5000)
			os.Exit(1)
		}
		json.Unmarshal(jsonstring, &a)
	}
	log.Debug("login URL: " + loginUrl)
	client := http.Client{Jar: jar}
	res, _ := client.PostForm(
		loginUrl,
		url.Values{"mail": {a.Mail}, "password": {a.Password}},
	)
	log.Tracef("%#v", res)
	log.Debug(res.Status)
	if res.StatusCode != 301 {
		log.Info("login failed")
		sleep(5000)
		os.Exit(1)
	} else {
		log.Info("login success")
	}
}

func download(videoId string, o Option) {
	log.Info("===================================================")
	log.Info("videoId: " + videoId)
	sleep(15000)

	// flv保管情報取得
	flvInfo := getFlvInfo(getFlvUrl + videoId)
	log.Tracef("%#v", flvInfo)

	// 動画情報取得(未ログインでも取得できる)
	nicovideo := getNicovideoThumbResponse(getThumbinfoUrl + videoId)

	// nicovideoThumbResponseが正常に取得できなかった場合の処理
	if nicovideo.Error.Code != "" {
		log.Warn(nicovideo.Error.Code + " " + nicovideo.Error.Description)
		return
	}
	// flvInfoが正常に取得できなかった場合の処理
	if flvInfo.Url == "" {
		log.Warn("flvInfo.Url is EMPTY.無料期間終了か、元から有料っぽい")
		return
	} else if strings.Contains(flvInfo.Url, "rtmpe://") {
		log.Warn("flvInfo.Url is Real Time Messaging Protocol.パトロールが難しいタイプの動画")
		saveRtmp(flvInfo, nicovideo)
		return
	}

	// ファイル名に/が入っている場合は消す
	title := strings.Replace(nicovideo.Thumb.Title, "/", "", -1)
	chName := nicovideo.Thumb.ChName
	nickname := nicovideo.Thumb.UserNickname
	movieType := "." + nicovideo.Thumb.MovieType
	txtType := ".txt"
	xmlType := ".xml"
	sizeHigh := nicovideo.Thumb.SizeHigh

	log.Info("target: " + title)

	filepath := *o.Destination
	if chName == "" {
		filepath = filepath + "/user/" + nickname
	} else {
		filepath = filepath + "/channel/" + chName
	}
	dest := filepath + "/" + title

	fi, _ := os.Stat(dest + movieType)
	fullsize := parseInt(sizeHigh)
	if fi == nil {
		log.Info("new download: " + dest)
	} else {
		if int(fi.Size()) == fullsize {
			log.Warn("video is already EXIST.動画は既に存在している: " + dest)

			// 存在しているが、-fオプションが付いている場合は動画情報とコメントを取りに行く
			if *o.IsForce {
				log.Info("force download.動画情報とコメント強制ダウンロード")
				saveNicovideoInfo(dest+txtType, nicovideo)
				saveComment(dest+xmlType, flvInfo)
			}

			return
		} else {
			log.Warn("redownload: " + dest)
			os.Remove(dest + movieType)
		}
	}

	log.Info("make dir: " + filepath)
	os.MkdirAll(filepath, 0711)

	saveNicovideoInfo(dest+txtType, nicovideo)
	saveComment(dest+xmlType, flvInfo)
	saveVideo(dest+movieType, flvInfo.Url, nicovideo, o)
}

// 動画情報ファイル出力
func saveNicovideoInfo(filePath string, nicovideo NicovideoThumbResponse) {
	buf, _ := xml.MarshalIndent(nicovideo, "", "  ")
	writeFile(filePath, buf)
}

// コメント取得
func saveComment(filePath string, flvInfo FlvInfo) {
	comment := getComment(flvInfo)
	writeFile(filePath, comment)
}

// 動画ファイル書き込み
func saveVideo(filepath string, videoUrl string, nicovideo NicovideoThumbResponse, o Option) {
	client := http.Client{Jar: jar}
	watchUrl := nicovideo.Thumb.WatchUrl
	size := parseInt(nicovideo.Thumb.SizeHigh)

	// videoUrlにアクセスする前にいったんwatchUrlをgetしてCookieを取得している必要がある
	// http://n-yagi.0r2.net/script/2009/12/nico2downloader.html
	// http://sekai.hatenablog.jp/entry/2013/02/26/164500
	client.Get(watchUrl)

	videoUrlDecode, _ := url.QueryUnescape(videoUrl)
	log.Debug("video server URL: " + videoUrlDecode)
	res, _ := client.Get(videoUrlDecode)
	log.Tracef("%#v", res)
	log.Debug(res.Status)
	defer res.Body.Close()

	log.Info("download: " + filepath)
	// File open
	file, _ := os.OpenFile(filepath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
	defer file.Close()

	copyFile(size, file, res.Body, o)
}

func saveRtmp(flvInfo FlvInfo, nicovideo NicovideoThumbResponse) {
	parseUrl, _ := url.Parse(flvInfo.Url)
	log.Tracef("%#v", parseUrl)
	host := parseUrl.Host
	tcUrl := parseUrl.Scheme + "://" + parseUrl.Host + parseUrl.Path
	playpath := parseUrl.RawQuery

	buf := strings.Split(flvInfo.Fmst, ":")
	fmst1 := buf[1]
	fmst2 := buf[0]

	pageUrl := nicovideo.Thumb.WatchUrl

	swfUrl := "http://res.nimg.jp/swf/player/secure_nccreator.swf?t=201111091500"
	flashVer := "WIN 11,6,602,180"

	log.Info(host)
	log.Info(tcUrl)
	log.Info(playpath)
	log.Info(pageUrl)
	log.Info(fmst1)
	log.Info(fmst2)
	log.Info(swfUrl)
	log.Info(flashVer)
}
