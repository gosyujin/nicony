# Nicony

![](./img/logo.png)

Nicony is a tool able to cache the video and the comment.

Powerd by Golang.

## Required

- account.json(Mail, Password)

## Usage

```golang
$ go run nicony.go
[nicony] 2016-01-09T01:45:53+09:00 [nicony.go:main:159] [INFO] nicony ver.0.3
[nicony] 2016-01-09T01:45:53+09:00 [nicony.go:login:229] [INFO] read ./account.json
[nicony] 2016-01-09T01:46:10+09:00 [nicony.go:main:169] [INFO] ===================================================
[nicony] 2016-01-09T01:51:17+09:00 [nicony.go:main:182] [INFO] target: 石上静香と東山奈央の英雄譚ＲＡＤＩＯ　第十四回
[nicony] 2016-01-09T01:51:17+09:00 [nicony.go:main:200] [INFO] new download: dest/channel/落第騎士の英雄譚（キャバルリィ）/石上静香と東山奈央の英雄譚ＲＡＤＩＯ　第十四回
[nicony] 2016-01-09T01:51:17+09:00 [nicony.go:main:211] [INFO] make dir: dest/channel/落第騎士の英雄譚（キャバルリィ）
[nicony] 2016-01-09T01:51:17+09:00 [nicony.go:write:504] [INFO] write file: dest/channel/落第騎士の英雄譚（キャバルリィ）/石上静香と東山奈央の英雄譚ＲＡＤＩＯ　第十四回.txt
[nicony] 2016-01-09T01:51:19+09:00 [nicony.go:write:504] [INFO] write file: dest/channel/落第騎士の英雄譚（キャバルリィ）/石上静香と東山奈央の英雄譚ＲＡＤＩＯ　第十四回.xml
[nicony] 2016-01-09T01:51:19+09:00 [nicony.go:downloadVideo:471] [INFO] download: dest/channel/落第騎士の英雄譚（キャバルリィ）/石上静香と東山奈央の英雄譚ＲＡＤＩＯ　第十四回.mp4
39.52 MB / 96.70 MB [==============>---------------------] 40.86 % 3.87 MB/s 14s
```
