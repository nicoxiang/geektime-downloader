# geektime-downloader

geektime-downloader 支持下载专栏为 PDF 文档和下载视频课。

[![go report card](https://goreportcard.com/badge/github.com/nicoxiang/geektime-downloader "go report card")](https://goreportcard.com/report/github.com/nicoxiang/geektime-downloader)
[![MIT license](https://img.shields.io/badge/license-MIT-brightgreen.svg)](https://opensource.org/licenses/MIT)

## Usage

### Prerequisites

- Chrome installed

### Install form source

```bash
# Go 1.16+
go install github.com/nicoxiang/geektime-downloader@latest

# Go version < 1.16
go get -u github.com/nicoxiang/geektime-downloader@latest
```

### Download binary files

See release page

### Sample

```bash
## windows 为例
## 账号密码方式登录（常用）
> geektime-downloader.exe -u "phone number"
## cookie 方式登录
> geektime-downloader.exe --gcid "gcid" --gcess "gcess"
```

### Help

```bash
## windows 为例
> geektime-downloader.exe -h

Geektime-downloader is used to download geek time lessons

Usage:
  geektime-downloader [flags]

Flags:
  -c, --concurrency int   下载并发数 (default 4)
  -f, --folder string     专栏和视频课的下载目标位置 (default "")
      --gcess string      极客时间 cookie 值 gcess
      --gcid string       极客时间 cookie 值 gcid
  -h, --help              help for geektime-downloader
  -u, --phone string      你的极客时间账号(手机号)
  -q, --quality string    下载视频清晰度(ld标清,sd高清,hd超清) (default "sd")
```

## Note

1. 文件下载目标位置可以通过 help 查看

2. Ctrl + C 退出程序

3. 如果选择下载所有后中断程序，可重新进入程序继续下载

4. 通过密码登录的情况下，为了避免多次登录账户，会在目录 [UserConfigDir](https://pkg.go.dev/os#UserConfigDir)/geektime-downloader 下存放用户的登录 cookie，如果不是在自己的电脑上执行，请在使用完毕程序后手动删除

## Inspired by 

* [geektime-dl](https://github.com/mmzou/geektime-dl)
