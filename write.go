package main

import (
	"github.com/cheggaaa/pb"
	log "github.com/cihub/seelog"
	"io"
	"io/ioutil"
	"os"
)

func writeFile(filepath string, body []byte) {
	log.Info("write file: " + filepath)

	ioutil.WriteFile(filepath, body, os.ModePerm)
}

func copyFile(size int, filepath io.Writer, source io.Reader, o Option) {
	log.Info("download start")

	sleep(15000)
	if *o.IsProgressBar {
		// プログレスバー
		progressBar := pb.New(size)
		progressBar.SetUnits(pb.U_BYTES)
		progressBar.SetRefreshRate(getDuration(10))
		progressBar.ShowCounters = true
		progressBar.ShowTimeLeft = true
		progressBar.ShowSpeed = true
		progressBar.SetMaxWidth(80)
		progressBar.Start()

		writer := io.MultiWriter(filepath, progressBar)
		io.Copy(writer, source)
		progressBar.FinishPrint("download complete")
	} else {
		writer := io.MultiWriter(filepath)
		io.Copy(writer, source)
	}
	log.Info("download complete")
	sleep(15000)
}
