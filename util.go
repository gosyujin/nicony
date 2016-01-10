package main

import (
	log "github.com/cihub/seelog"
	"strconv"
	"time"
)

func sleep(minutes int64) {
	time.Sleep(time.Millisecond * time.Duration(minutes))
}

func getDuration(minutes int64) time.Duration {
	return time.Millisecond * time.Duration(minutes)
}

func parseInt(s string) int {
	i, err := strconv.Atoi(s)
	if err != nil {
		log.Error("parse error: " + s)
	}
	return i
}
