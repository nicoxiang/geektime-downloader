# geektime-downloader

geektime-downloader 支持下载专栏为 PDF/Markdown 文档和下载视频课。

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

See [release page](https://github.com/nicoxiang/geektime-downloader/releases)

### Sample

```bash
## Windows 为例
## Windows 推荐使用 Windows Terminal 打开

## 账号密码方式登录（常用）
> geektime-downloader.exe -u "phone number"
## cookie 方式登录
> geektime-downloader.exe --gcid "gcid" --gcess "gcess"
```

### Help

```bash
## Windows 为例
> geektime-downloader.exe -h

Geektime-downloader is used to download geek time lessons

Usage:
  geektime-downloader [flags]

Flags:
      --comments         是否需要专栏的第一页评论 (default true)
  -f, --folder string    专栏和视频课的下载目标位置 (default "")
      --gcess string     极客时间 cookie 值 gcess
      --gcid string      极客时间 cookie 值 gcid
  -h, --help             help for geektime-downloader
      --output int       专栏的输出内容(1pdf,2markdown,4audio)可自由组合 (default 1)
  -u, --phone string     你的极客时间账号(手机号)
  -q, --quality string   下载视频清晰度(ld标清,sd高清,hd超清) (default "sd")  
```

## Note

1. 文件下载目标位置可以通过 help 查看。默认情况下 Windows 位于 %USERPROFILE%/geektime-downloader 下；Unix, 包括 macOS, 位于 $HOME/geektime-downloader 下

2. 如何查看课程 ID?

打开极客时间[课程列表页](https://time.geekbang.org/resource)，选择你想要查看的课程，在新打开的课程详情 Tab 页，查看 URL 最后的数字，例如下面的链接中 100056701 就是课程 ID：

```
https://time.geekbang.org/column/intro/100056701
```

3. Ctrl + C 退出程序

4. 如何下载 Markdown 格式和文章音频?

默认情况下载专栏的输出内容只有 PDF，可以通过 --output 参数按需选择是否需要下载 Markdown 格式和文章音频。比如 --output 3 就是下载 PDF 和 Markdown；--output 6 就是下载 Markdown 和音频；--output 7 就是下载所有。

Markdown 格式虽然显示效果上不及 PDF，但优势为可以显示完整的代码块（PDF 代码块在水平方向太长时会有缺失）并保留了原文中的超链接。

5. 如果选择下载所有后中断程序，可重新进入程序继续下载

6. 通过密码登录的情况下，为了避免多次登录账户，会在目录 [UserConfigDir](https://pkg.go.dev/os#UserConfigDir)/geektime-downloader 下存放用户的登录 cookie，如果不是在自己的电脑上执行，请在使用完毕程序后手动删除

## Inspired by 

* [geektime-dl](https://github.com/mmzou/geektime-dl)
