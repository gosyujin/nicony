package main

import (
	"encoding/json"
	log "github.com/cihub/seelog"
	"io/ioutil"
	"net/http"
	"net/url"
)

// 指定された動画のFLV保管URLの情報 http://dic.nicovideo.jp/a/ニコニコ動画api
type FlvInfo struct {
	ThreadId         string   `json:"thread_id"`          //1 コメントDLで使う
	L                string   `json:"l"`                  //2 コメントDLで使う、60で割って+1して使う
	Url              string   `json:"url"`                //3 動画DLで使う
	Ms               string   `json:"ms"`                 //4 コメントDLで使う
	MsSub            string   `json:"ms_sub"`             //5
	UserId           string   `json:"user_id"`            //6 コメントDLで使う
	IsPremium        string   `json:"is_premium"`         //7 (プレミアムなら1)
	Nickname         string   `json:"nickname"`           //8
	Time             []string `json:"time"`               //9
	Done             string   `json:"done"`               //10
	NgRv             string   `json:"ng_rv"`              //11
	Hms              string   `json:"hms"`                //12
	Hmsp             string   `json:"hmsp"`               //13
	Hmst             string   `json:"hmst"`               //14
	Hmstk            string   `json:"hmstk"`              //15
	UserKey          string   `json:"user_key"`           //16
	NeedsKey         string   `json:"needs_key"`          //17 公式放送のみ存在？
	OptionalThreadId string   `json:"optional_thread_id"` //18 公式放送のみ存在？
	NgCh             string   `json:"ng_ch"`              //19 公式放送のみ存在？
	Closed           string   `json:"closed"`             //20
	Deleted          string   `json:"deleted"`            //21
	Error            string   `json:"error"`              //22
}

func getFlvInfo(getFlvUrl string) FlvInfo {
	client := http.Client{Jar: jar}

	log.Debug("get getFlvUrl " + getFlvUrl)
	res, _ := client.Get(getFlvUrl)
	body, _ := ioutil.ReadAll(res.Body)
	log.Trace(string(body))
	log.Debug(res.Status)
	defer res.Body.Close()

	// クエリパラメータをいったんmap -> json -> 構造体フィールドタグに落としこむ
	queryMap, _ := url.ParseQuery(string(body))
	log.Tracef("%#v", queryMap)

	j, _ := json.Marshal(queryMap)
	log.Tracef("%#v", string(j))

	flvInfo := FlvInfo{}
	err := json.Unmarshal(j, &flvInfo)
	if err != nil {
		log.Error(err)
	}
	log.Tracef("%#v", flvInfo)

	//TODO 必須パラメータ存在チェック
	//log.Debug(f.ThreadId)
	//log.Debug(f.Url)
	//log.Debug(f.Ms)
	//log.Debug(f.UserId)

	return flvInfo
}
