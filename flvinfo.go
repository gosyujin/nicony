package main

import (
	log "github.com/cihub/seelog"
	"io/ioutil"
	"net/http"
	"net/url"
)

// 指定された動画のFLV保管URLの情報 http://dic.nicovideo.jp/a/ニコニコ動画api
type FlvInfo struct {
	ThreadId            string //1 コメントDLで使う
	L                   string //2 コメントDLで使う、60で割って+1して使う
	Url                 string //3 動画DLで使う
	Ms                  string //4 コメントDLで使う
	MsSub               string //5
	UserId              string //6 コメントDLで使う
	IsPremium           string //7 (プレミアムなら1)
	Nickname            string //8
	Time                string //9
	Done                string //10
	NgRv                string //11
	Hms                 string //12
	Hmsp                string //13
	Hmst                string //14
	Hmstk               string //15
	UserKey             string //16
	NeedsKey            string //17 公式放送のみ存在？
	OptionalThreadId    string //18 公式放送のみ存在？
	NgCh                string //19 公式放送のみ存在？
	Closed              string //20
	Deleted             string //21
	Error               string //22
	Fmst                string //23
	NotAvailablePostkey string //24
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
	queryMap, _ := url.ParseQuery(string(body))
	for key, v := range queryMap {
		value := v[0]

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
		case "closed":
			// 1だと不正終了っぽい
			f.Closed = value
		case "deleted":
			// 2だと削除された動画っぽい
			f.Deleted = value
		case "error":
			// invalid_v1だと有料放送っぽい？
			f.Error = value
		case "fmst":
			f.Fmst = value
		case "not_available_postkey":
			f.NotAvailablePostkey = value
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
