package cmd

import (
	"context"
	"os"
	"os/signal"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/nicoxiang/geektime-downloader/internal/config"
	"github.com/nicoxiang/geektime-downloader/internal/fsm"
	"github.com/nicoxiang/geektime-downloader/internal/geektime"
)

var (
	geektimeClient *geektime.Client
	cfg            config.AppConfig
)

func init() {
	userHomeDir, _ := os.UserHomeDir()
	defaultDownloadFolder := filepath.Join(userHomeDir, config.GeektimeDownloaderFolder)

	rootCmd.Flags().StringVar(&cfg.Gcid, "gcid", "", "极客时间 cookie 值 gcid")
	rootCmd.Flags().StringVar(&cfg.Gcess, "gcess", "", "极客时间 cookie 值 gcess")
	rootCmd.Flags().StringVarP(&cfg.DownloadFolder, "folder", "f", defaultDownloadFolder, "专栏和视频课的下载目标位置")
	rootCmd.Flags().StringVarP(&cfg.Quality, "quality", "q", "sd", "下载视频清晰度(ld标清,sd高清,hd超清)")
	rootCmd.Flags().IntVar(&cfg.DownloadComments, "comments", 1, "是否下载评论(0不下载,1下载首页评论,2下载所有评论)")
	rootCmd.Flags().IntVar(&cfg.ColumnOutputType, "output", 1, "专栏的输出内容(1pdf,2markdown,4audio)可自由组合")
	rootCmd.Flags().IntVar(&cfg.PrintPDFWaitSeconds, "print-pdf-wait", 5, "Chrome生成PDF前的等待页面加载时间, 单位为秒, 默认5秒")
	rootCmd.Flags().IntVar(&cfg.PrintPDFTimeoutSeconds, "print-pdf-timeout", 60, "Chrome生成PDF的超时时间, 单位为秒, 默认60秒")
	rootCmd.Flags().IntVar(&cfg.Interval, "interval", 1, "下载资源的间隔时间, 单位为秒, 默认1秒")
	rootCmd.Flags().BoolVar(&cfg.IsEnterprise, "enterprise", false, "是否下载企业版极客时间资源")

	rootCmd.MarkFlagsRequiredTogether("gcid", "gcess")
}

var rootCmd = &cobra.Command{
	Use:          "geektime-downloader",
	Short:        "Geektime-downloader is used to download geek time lessons",
	SilenceUsage: true,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return config.ValidateConfig(&cfg)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		readCookies := config.ReadCookiesFromInput(&cfg)

		geektimeClient = geektime.NewClient(readCookies)

		runner := fsm.NewFSMRunner(cmd.Context(), &cfg, geektimeClient)
		return runner.Run()
	},
}

// Execute ...
func Execute() {
	ctx := context.Background()

	ctx, cancel := context.WithCancel(ctx)
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	defer func() {
		signal.Stop(c)
		cancel()
	}()
	go func() {
		select {
		case <-c:
			cancel()
		case <-ctx.Done():
		}
	}()

	if err := rootCmd.ExecuteContext(ctx); err != nil {
		os.Exit(1)
	}
}
