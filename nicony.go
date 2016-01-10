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
	getNicorepoUrl  = "http://www.nicovideo.jp/my/top/all?innerPage=1&mode=next_page"
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
	IsAnsi        *bool   // ログ出力をAnsiカラーにするか
	IsProgressBar *bool   // ダウンロード時プログレスバーを表示するか
	IsVersion     *bool   // バージョン表示
	LogLevel      *string // ログレベル
	VideoId       *string // ビデオID
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
		// 引数に指定された動画取得
		download(*o.VideoId, o)
	}

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

func download(url string, o Option) {
	log.Info("===================================================")
	sleep(15000)

	// flv保管情報取得
	flvInfo := getFlvInfo(getFlvUrl + url)
	log.Tracef("%#v", flvInfo)

	// 動画情報取得(未ログインでも取得できる)
	nicovideo := getNicovideoThumbResponse(getThumbinfoUrl + url)

	title := strings.Replace(nicovideo.Thumb.Title, "/", "", -1)
	chName := nicovideo.Thumb.ChName
	nickname := nicovideo.Thumb.UserNickname
	movieType := "." + nicovideo.Thumb.MovieType
	txtType := ".txt"
	xmlType := ".xml"
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

	fi, _ := os.Stat(dest + movieType)
	fullsize := parseInt(sizeHigh)
	if fi == nil {
		log.Info("new download: " + dest)
	} else {
		if int(fi.Size()) == fullsize {
			log.Warn("video is already EXIST.動画は既に存在している: " + dest)
			return
		} else {
			log.Warn("redownload: " + dest)
			os.Remove(dest + movieType)
		}
	}

	log.Info("make dir: " + filepath)
	os.MkdirAll(filepath, 0711)

	// 動画情報ファイル出力
	buf, _ := xml.MarshalIndent(nicovideo, "", "  ")
	writeFile(dest+txtType, buf)

	// コメント取得
	comment := getComment(flvInfo)
	writeFile(dest+xmlType, comment)

	// 動画ファイル書き込み
	saveVideo(dest+movieType, flvInfo.Url, nicovideo, o)
}

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
	log.Debug(res.Status)
	defer res.Body.Close()

	log.Info("download: " + filepath)
	// File open
	file, _ := os.OpenFile(filepath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
	defer file.Close()

	copyFile(size, file, res.Body, o)
}
