package cmd

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"time"

	"github.com/manifoldco/promptui"
	"github.com/nicoxiang/geektime-downloader/internal/config"
	"github.com/nicoxiang/geektime-downloader/internal/fsm"
	"github.com/nicoxiang/geektime-downloader/internal/geektime"
	"github.com/spf13/cobra"
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
	rootCmd.Flags().BoolVar(&cfg.DownloadComments, "comments", true, "是否需要专栏的第一页评论")
	rootCmd.Flags().IntVar(&cfg.ColumnOutputType, "output", 1, "专栏的输出内容(1pdf,2markdown,4audio)可自由组合")
	rootCmd.Flags().IntVar(&cfg.PrintPDFWaitSeconds, "print-pdf-wait", 8, "Chrome生成PDF前的等待页面加载时间, 单位为秒, 默认8秒")
	rootCmd.Flags().IntVar(&cfg.PrintPDFTimeoutSeconds, "print-pdf-timeout", 60, "Chrome生成PDF的超时时间, 单位为秒, 默认60秒")
	rootCmd.Flags().IntVar(&cfg.Interval, "interval", 1, "下载资源的间隔时间, 单位为秒, 默认1秒")
	rootCmd.Flags().BoolVar(&cfg.IsEnterprise, "enterprise", false, "是否下载企业版极客时间资源")

	rootCmd.MarkFlagsRequiredTogether("gcid", "gcess")
}

var rootCmd = &cobra.Command{
	Use:           "geektime-downloader",
	Short:         "Geektime-downloader is used to download geek time lessons",
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		if cfg.Quality != "ld" && cfg.Quality != "sd" && cfg.Quality != "hd" {
			return fmt.Errorf("argument 'quality' is not valid")
		}
		if cfg.ColumnOutputType <= 0 || cfg.ColumnOutputType >= 8 {
			return fmt.Errorf("argument 'output' is not valid")
		}
		var readCookies []*http.Cookie
		if cfg.Gcid != "" && cfg.Gcess != "" {
			readCookies = readCookiesFromInput()
		} else {
			return fmt.Errorf("argument 'gcid' or 'gcess' is not valid")
		}

		geektimeClient = geektime.NewClient(readCookies)

		runner := fsm.NewFSMRunner(cmd.Context(), &cfg, geektimeClient)
		err := runner.Run()
		if errors.Is(err, context.Canceled) || errors.Is(err, promptui.ErrInterrupt) {
			// 用户中断，已经输出提示，不要再让 Cobra 打 help
			return nil
		}
		return err
	},
}

func readCookiesFromInput() []*http.Cookie {
	oneyear := time.Now().Add(180 * 24 * time.Hour)
	cookies := make([]*http.Cookie, 2)
	m := make(map[string]string, 2)
	m[geektime.GCID] = cfg.Gcid
	m[geektime.GCESS] = cfg.Gcess
	c := 0
	for k, v := range m {
		cookies[c] = &http.Cookie{
			Name:     k,
			Value:    v,
			Domain:   geektime.GeekBangCookieDomain,
			HttpOnly: true,
			Expires:  oneyear,
		}
		c++
	}
	return cookies
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
