package main

import (
	"encoding/json"
	log "github.com/cihub/seelog"
	set "github.com/deckarep/golang-set"
	"io/ioutil"
	"net/http"
	_ "sort"
	"strings"
)

type Nicorepo struct {
	Meta   Meta     `json:"meta"`
	Data   []Data   `json:"data"`
	Errors []string `json:"errors"`
	Status string   `json:"status"`
}

type Meta struct {
	Status         int    `json:"status"`
	MaxId          string `json:"maxId"`
	MinId          string `json:"minId"`
	ImpressionId   string `json:"impressionId"`
	ClientAppGroup string `json:"clientAppGroup"`
	Limit          int    `json:"_limit"`
}

type Data struct {
	Id string `json:"id"`
	//	Topic         string  `json:"topic"`
	//	CreatedAt     string  `json:"createdAt"`
	//	IsVisible     string  `json:"isVisible"`
	//	IsMuted       string  `json:"isMuted"`
	//	IsDeletable   string  `json:"isDeletable"`
	//	MuteContext   string  `json:"muteContext"`
	//	SenderChannel string  `json:"senderChannel"`
	//	ActionLog     string  `json:"actionLog"`
	Program Program `json:"program"`
	Video   Video   `json:"video"`
}

type Program struct {
	Id string `json:"id"`
	//	BeginAt      string `json:"beginAt"`
	//	IsPayProgram string `json:"isPayProgram"`
	//	ThumbnailUrl string `json:"thumbnailUrl"`
	Title string `json:"title"`
}

type Video struct {
	Id string `json:"id"`
	//Resolution       string `json:"resolution"`
	//Status           string `json:"status"`
	//ThumbnailUrl     string `json:"thumbnailUrl"`
	Title string `json:"title"`
	//VideoWatchPageId string `json:"videoWatchPageId"`
}

// ニコレポから動画IDのリストを取得する
func getNicorepo(getNicorepoUrl string, cursor string, links []string) []string {
	client := http.Client{Jar: jar}
	log.Trace("getNicorepo URL: " + getNicorepoUrl + cursor)

	res, _ := client.Get(getNicorepoUrl + cursor)
	body, _ := ioutil.ReadAll(res.Body)
	defer res.Body.Close()

	n := Nicorepo{}
	jsonstring, _ := ioutil.ReadAll(strings.NewReader(string(body)))
	json.Unmarshal(jsonstring, &n)

	for _, data := range n.Data {
		if data.Program.Id != "" {
			links = append(links, data.Program.Id)
		}
		if data.Video.Id != "" {
			links = append(links, data.Video.Id)
		}

	}

	if n.Meta.MinId == "" {
		return uniq(links)
	} else {
		return getNicorepo(getNicorepoUrl, "&cursor="+n.Meta.MinId, links)
	}
}

func uniq(links []string) []string {
	//sort.Strings(links)
	log.Trace(len(links))
	log.Trace(links)

	// golang-setのNewSetFromSliceを使って重複削除
	// http://ashitani.jp/golangtips/tips_slice.html#slice_Uniq
	// https://github.com/golang/go/wiki/InterfaceSlice
	var interfaceSlice []interface{} = make([]interface{}, len(links))
	for _, d := range links {
		interfaceSlice = append(interfaceSlice, d)
	}
	slice := set.NewSetFromSlice(interfaceSlice)

	var uniqueLinks = []string{}
	for _, d := range slice.ToSlice() {
		s, ok := d.(string)
		if ok {
			uniqueLinks = append(uniqueLinks, s)
		}
	}

	//sort.Strings(uniqueLinks)
	log.Debug(len(uniqueLinks))
	log.Debug(uniqueLinks)

	return uniqueLinks
}
