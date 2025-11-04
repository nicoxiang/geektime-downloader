package config

const (
	// GeektimeDownloaderFolder app config and download root dolder name
	GeektimeDownloaderFolder = "geektime-downloader"
)

type AppConfig struct {
	Gcid                   string
	Gcess                  string
	DownloadFolder         string
	Quality                string
	DownloadComments       bool
	ColumnOutputType       int
	PrintPDFWaitSeconds    int
	PrintPDFTimeoutSeconds int
	Interval               int
	IsEnterprise           bool
}
