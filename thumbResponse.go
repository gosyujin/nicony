package main

import (
	"encoding/xml"
	log "github.com/cihub/seelog"
	"io/ioutil"
	"net/http"
)

// 指定された動画の動画情報 http://dic.nicovideo.jp/a/ニコニコ動画api
type NicovideoThumbResponse struct {
	Thumb Thumb `xml:"thumb"`
	Error Error `xml:"error"`
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

type Error struct {
	Code        string `xml:"code"`
	Description string `xml:"description"`
}

func getNicovideoThumbResponse(getThumbinfoUrl string) NicovideoThumbResponse {
	log.Debug("get getNicovideoThumbResponse URL: " + getThumbinfoUrl)

	res, _ := http.Get(getThumbinfoUrl)
	log.Debug(res.Status)
	body, _ := ioutil.ReadAll(res.Body)
	defer res.Body.Close()

	nicovideo := NicovideoThumbResponse{}
	xml.Unmarshal(body, &nicovideo)
	log.Tracef("%#v", nicovideo)

	return nicovideo
}
