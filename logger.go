package main

import (
	log "github.com/cihub/seelog"
)

func initLogger(o Option) {
	var logConfig string
	var logLevel string

	switch *o.LogLevel {
	case "critical":
		logLevel = "critical"
	case "error":
		logLevel = "error,critical"
	case "warn":
		logLevel = "warn,error,critical"
	case "info":
		logLevel = "info,warn,error,critical"
	case "debug":
		logLevel = "debug,info,warn,error,critical"
	case "trace":
		logLevel = "trace,debug,info,warn,error,critical"
	default:
		logLevel = "debug,info,warn,error,critical"
	}

	// Ansiカラー適用する場合としない場合
	if *o.IsAnsi {
		logConfig = `
		  <seelog type="adaptive" mininterval="200000000" maxinterval="1000000000" critmsgcount="5">
		    <formats>
		      <format id="console" format="%EscM(36)[nicony]%EscM(39) %EscM(32)%Date(2006-01-02T15:04:05Z07:00)%EscM(39) %EscM(33)[%File:%FuncShort:%Line]%EscM(39) %EscM(46)[%LEVEL]%EscM(49) %Msg%n" />
		      <format id="plane" format="[nicony] %Date(2006-01-02T15:04:05Z07:00) [%File:%FuncShort:%Line] [%LEVEL] %Msg%n" />
		    </formats>
		    <outputs>
		      <filter formatid="console" levels="` + logLevel + `">
		        <console />
		      </filter>
		      <filter formatid="plane" levels="trace,debug,info,warn,error,critical">
		        <rollingfile filename="./var/log.txt" type="size" maxsize="1024000" maxrolls="500" />
		      </filter>
		    </outputs>
		  </seelog>`
	} else {
		logConfig = `
		  <seelog type="adaptive" mininterval="200000000" maxinterval="1000000000" critmsgcount="5">
		    <formats>
		      <format id="plane" format="[nicony] %Date(2006-01-02T15:04:05Z07:00) [%File:%FuncShort:%Line] [%LEVEL] %Msg%n" />
		    </formats>
		    <outputs>
		      <filter formatid="plane" levels="` + logLevel + `">
		        <console />
		      </filter>
		      <filter formatid="plane" levels="trace,debug,info,warn,error,critical">
		        <rollingfile filename="./var/log.txt" type="size" maxsize="1024000" maxrolls="500" />
		      </filter>
		    </outputs>
		  </seelog>`
	}

	logger, _ := log.LoggerFromConfigAsBytes([]byte(logConfig))
	log.ReplaceLogger(logger)
}
