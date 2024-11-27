# geektime-downloader

geektime-downloader 支持下载以下极客时间网站资源。

**极客时间**
- [x] 专栏(PDF/Markdown/音频)
- [x] 视频课
- [x] 每日一课
- [x] 大厂案例
- [x] 训练营视频
- [ ] 线下大会

**企业版极客时间**
- [ ] 体系课
- [ ] 每日一课
- [ ] 大厂案例
- [ ] 生态课
- [x] 训练营视频

部分资源暂未支持下载，欢迎PR。


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
      --comments                是否需要专栏的第一页评论 (default true)
      --enterprise              是否下载企业版极客时间资源
  -f, --folder string           专栏和视频课的下载目标位置 (default "C:\\Users\\nico\\geektime-downloader")
      --gcess string            极客时间 cookie 值 gcess
      --gcid string             极客时间 cookie 值 gcid
  -h, --help                    help for geektime-downloader
      --interval int            下载资源的间隔时间, 单位为秒, 默认1秒 (default 1)
      --output int              专栏的输出内容(1pdf,2markdown,4audio)可自由组合 (default 1)
      --print-pdf-timeout int   Chrome生成PDF的超时时间, 单位为秒, 默认60秒 (default 60)
      --print-pdf-wait int      Chrome生成PDF前的等待页面加载时间, 单位为秒, 默认8秒 (default 8)
  -q, --quality string          下载视频清晰度(ld标清,sd高清,hd超清) (default "sd")
```

## Note

### 文件下载目标位置

文件下载目标位置可以通过 help 查看。默认情况下 Windows 位于 %USERPROFILE%/geektime-downloader 下；Unix, 包括 macOS, 位于 $HOME/geektime-downloader 下

### 如何查看课程 ID?

**普通课程：**

打开极客时间[课程列表页](https://time.geekbang.org/resource)，选择你想要查看的课程，在新打开的课程详情 Tab 页，查看 URL 最后的数字，例如下面的链接中 100056701 就是课程 ID：

```
https://time.geekbang.org/column/intro/100056701
```

**训练营课程：**

打开极客时间[训练营课程列表页](https://u.geekbang.org/schedule)，选择你想要查看的课程，在新打开的课程详情 Tab 页，查看 URL ```lesson/```后的数字，例如下面的链接中 419 就是课程 ID：

```
https://u.geekbang.org/lesson/419?article=535616
```

**每日一课课程：**

选择你想要下载的视频，查看 URL ```dailylesson/detail/```后的数字，例如下面的链接中 100122405 就是课程 ID：

```
https://time.geekbang.org/dailylesson/detail/100122405
```

**大厂案例课程：**

选择你想要下载的视频，查看 URL ```qconplus/detail/```后的数字，例如下面的链接中 100110494 就是课程 ID：

```
https://time.geekbang.org/qconplus/detail/100110494
```

**公开课课程：**

选择你想要下载的视频，查看 URL ```opencourse/intro/``` 或 ```opencourse/videointro/```后的数字，例如下面的链接中 100546701 就是课程 ID：

```
https://time.geekbang.org/opencourse/videointro/100546701
```

**其他：**

打开极客时间[我的课程-其他](https://time.geekbang.org/dashboard/course)，选择你想要查看的课程，在新打开的课程详情 Tab 页，查看 URL ```course/intro/``` 最后的数字，例如下面的链接中 100551201 就是课程 ID：

```
https://time.geekbang.org/course/intro/100551201
```

**企业版训练营：**

选择你想要查看的课程，查看 URL ```mall/product/```后的数字，例如下面的链接中 100618109 就是课程 ID：

```
https://b.geekbang.org/mall/product/100618109
```

### 为什么我下载的PDF是空白页?
首先下载课程请保证VPN已关闭。在此前提下如果仍然出现空白页情况，说明后台Chrome网页加载速度较慢，可以尝试加大--print-pdf-wait参数，保证页面完全加载完成后再开始生成PDF。

### 为什么我下载PDF一直提示超时?
首先下载课程请保证VPN已关闭。在此前提下如果下载持续出现超时，有可能是因为课程章节图片等内容较多，生成速度慢，比如课程《AI 绘画核心技术与实战》中的部分章节，可以尝试加大--print-pdf-timeout参数，并耐心等待。

### 如何下载专栏的 Markdown 格式和文章音频?

默认情况下载专栏的输出内容只有 PDF，可以通过 --output 参数按需选择是否需要下载 Markdown 格式和文章音频。比如 --output 3 就是下载 PDF 和 Markdown；--output 6 就是下载 Markdown 和音频；--output 7 就是下载所有。

Markdown 格式虽然显示效果上不及 PDF，但优势为可以显示完整的代码块（PDF 代码块在水平方向太长时会有缺失）并保留了原文中的超链接。

现在部分新课程的专栏文章中会包含视频，如课程《Kubernetes 入门实战课》等，目前程序会自动下载文章所包含的视频，视频目录在文章所在目录的子目录 videos 下，此类文章PDF的下载会耗费更多时间，请耐心等待。

### 退出程序和继续下载

Ctrl + C 退出程序。如果选择“下载所有”后中断程序，可重新进入程序继续下载。

### 隐私相关

通过密码登录的情况下，为了避免多次登录账户，会在目录 [UserConfigDir](https://pkg.go.dev/os#UserConfigDir)/geektime-downloader 下存放用户的登录 cookie，如果不是在自己的电脑上执行，建议在使用完毕程序后手动删除
