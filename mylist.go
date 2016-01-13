package main

import (
	"encoding/xml"
	log "github.com/cihub/seelog"
	"io/ioutil"
	"net/http"
	"strings"
)

// マイリストのrss
type Rss struct {
	Channel Channel `xml:"channel"`
}

type Channel struct {
	Title         string `xml:"title"`
	Link          string `xml:"link"`
	Atom          string `xml:"atom"`
	Description   string `xml:"description"`
	PubDate       string `xml:"pubDate"`
	LastBuildDate string `xml:"lastBuildDate"`
	Generator     string `xml:"generator"`
	Dc            string `xml:"dc:creator"`
	Language      string `xml:"language"`
	Copyright     string `xml:"copyright"`
	Docs          string `xml:"docs"`
	Item          []Item `xml:"item"`
}

type Item struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Guid        string `xml:"guid"`
	PubDate     string `xml:"pubDate"`
	Description string `xml:"description"`
}

func getMylist(getMylistUrl string) []string {
	log.Debug("get getMylist URL: " + getMylistUrl)

	res, _ := http.Get(getMylistUrl)
	log.Tracef("%#v", res)
	log.Debug(res.Status)
	body, _ := ioutil.ReadAll(res.Body)
	defer res.Body.Close()

	rss := Rss{}
	xml.Unmarshal(body, &rss)
	log.Tracef("%#v", rss)

	// TODO この3つがうまくとれない
	// <link>http://www.nicovideo.jp/mylist/12346423</link>
	// <atom:link rel="self" type="application/rss+xml" href="http://www.nicovideo.jp/mylist/12346423?rss=2.0"/>
	// <dc:creator>ゆーざ</dc:creator>

	var mylist []string
	for _, video := range rss.Channel.Item {
		buf := strings.Split(video.Link, "/")
		number := buf[len(buf)-1]

		mylist = append(mylist, number)
	}
	return mylist
}
