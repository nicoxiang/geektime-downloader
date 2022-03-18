# geektime-downloader

geektime-downloader 目前支持下载专栏为PDF文档。

## Usage

### Install form source

```bash
# Go 1.16+
go install github.com/nicoxiang/geektime-downloader

# Go version < 1.16
go get -u github.com/nicoxiang/geektime-downloader
```

### Sample

```bash
## windows 为例
> geektime-downloader.exe -u "phone number"
```

### Help

```bash
## windows 为例
> geektime-downloader.exe -h

Geektime-downloader is used to download geek time lessons

Usage:
  geektime-downloader [flags]

Flags:
  -c, --concurrency int   下载文章的并发数 (default 4)
  -f, --folder string     PDF 文件下载目标位置 (default "")
  -h, --help              help for geektime-downloader
  -u, --phone string      你的极客时间账号(手机号)(required)
```

## Note

为了避免多次登录账户，在目录 [UserConfigDir](https://pkg.go.dev/os#UserConfigDir)/geektime-downloader 下会存放用户的登录信息，如果不是在自己的电脑上执行，请在使用完毕程序后手动删除。

## Inspired by 

* [geektime-dl](https://github.com/mmzou/geektime-dl)
