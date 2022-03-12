# geektime-downloader

geektime-downloader 目前支持下载专栏为PDF文档。

## Usage

```bash
> geektime-downloader.exe -u "phone number"
```

```bash
> geektime-downloader.exe -h

Geektime-downloader is used to download geek time lessons

Usage:
  geektime-downloader [flags]

Flags:
  -c, --concurrency int   下载文章的并发数 (default 5)
  -h, --help              help for geektime-downloader
  -u, --phone string      你的极客时间账号(手机号)
```

## Inspired by 

* [geektime-dl](https://github.com/mmzou/geektime-dl)

## License

MIT