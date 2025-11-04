package main

import (
	"os"

	"github.com/nicoxiang/geektime-downloader/cmd"
)

func init() {
	// Get around rsa1024min panic issue
    os.Setenv("GODEBUG", os.Getenv("GODEBUG") + ",rsa1024min=0")
}

func main() {
	cmd.Execute()
}
