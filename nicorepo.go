package main

import (
	"github.com/PuerkitoBio/goquery"
	log "github.com/cihub/seelog"
	set "github.com/deckarep/golang-set"
	"io/ioutil"
	"net/http"
	"sort"
	"strings"
)

// ニコレポから動画IDのリストを取得する
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
			links = addList(s, links)
		})
		// 公式が動画投稿
		s.Find(".log-community-video-upload .log-target-info").Each(func(_ int, s *goquery.Selection) {
			links = addList(s, links)
		})
		// ユーザーが動画投稿
		s.Find(".log-user-video-upload .log-target-info").Each(func(_ int, s *goquery.Selection) {
			links = addList(s, links)
		})
		// xx再生を達成
		s.Find(".log-user-video-round-number-of-view-counter .log-target-info").Each(func(_ int, s *goquery.Selection) {
			links = addList(s, links)
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
		return uniq(links)
	} else {
		return getNicorepo(nicovideojpUrl+nextPageLink, links)
	}
}

func addList(s *goquery.Selection, links []string) []string {
	log.Debug(s.Find("a").Text())
	link, _ := s.Find("a").Attr("href")
	buf := strings.Split(link, "/")
	number := buf[len(buf)-1]

	links = append(links, number)
	return links
}

func uniq(links []string) []string {
	sort.Strings(links)
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

	sort.Strings(uniqueLinks)
	log.Debug(len(uniqueLinks))
	log.Debug(uniqueLinks)

	return uniqueLinks
}
