package main

import (
	"fmt"
	log "github.com/cihub/seelog"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

//コメントDLにPostに埋め込むxmlに使うパラメータ
type ThreadKeyInfo struct {
	ThreadKey string //1 コメントDLで使う、ユーザ動画は空、公式動画は値あり？
	Force184  string //2 コメントDLで使う、必ず1？
}

func getComment(flvInfo FlvInfo) []byte {
	threadKeyInfo := getThreadKeyInfo(flvInfo.ThreadId)
	sleep(5000)

	minutes := (parseInt(flvInfo.L) / 60) + 1
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
	res, err := client.Post(
		messageServer,
		"text/xml",
		strings.NewReader(packetXml),
	)
	if err != nil {
		log.Error(err)
	}
	log.Tracef("%#v", res)
	log.Debug(res.Status)

	body, _ := ioutil.ReadAll(res.Body)
	defer res.Body.Close()

	sleep(5000)
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
	queryMap, _ := url.ParseQuery(string(body))
	for key, v := range queryMap {
		value := v[0]

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
