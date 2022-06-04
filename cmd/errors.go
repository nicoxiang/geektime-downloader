package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/manifoldco/promptui"
	"github.com/nicoxiang/geektime-downloader/internal/config"
	"github.com/nicoxiang/geektime-downloader/internal/geektime"
	"github.com/nicoxiang/geektime-downloader/internal/pdf"
)

func checkError(err error) {
	if err != nil {
		// special newline case
		if errors.Is(err, pdf.ErrGeekTimeRateLimit) ||
			os.IsTimeout(err) {
			fmt.Println()
		}

		var eg *geektime.ErrGeekTimeAPIBadCode
		if errors.Is(err, context.Canceled) ||
			errors.Is(err, promptui.ErrInterrupt) {
			os.Exit(1)
		} else if errors.As(err, &eg) ||
			errors.Is(err, geektime.ErrWrongPassword) ||
			errors.Is(err, geektime.ErrTooManyLoginAttemptTimes) {
			exitWithMsg(err.Error())
		} else if errors.Is(err, pdf.ErrGeekTimeRateLimit) ||
			errors.Is(err, geektime.ErrAuthFailed) {
			exitWithMsg(err.Error())
			exitAndRemoveConfig(err)
		} else if os.IsTimeout(err) {
			exitWithMsg("请求超时")
		} else {
			fmt.Fprintf(os.Stderr, "An error occurred: %v\n", err.Error())
			os.Exit(1)
		}
	}
}

func exitAndRemoveConfig(err error) {
	if err := config.RemoveConfig(phone); err != nil {
		exitWithMsg(err.Error())
	}
	os.Exit(1)
}

func exitWithMsg(msg string) {
	fmt.Fprintln(os.Stderr, msg)
	os.Exit(1)
}
