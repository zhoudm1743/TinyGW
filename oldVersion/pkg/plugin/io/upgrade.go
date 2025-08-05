package io

import (
	"os"
)

func Upgrade(sourceFile, targetFile string) {
	if _, err := os.Stat(targetFile); err == nil {
		OsCommand("rm -rf " + targetFile)
	}
	DownloadFile(sourceFile, targetFile)
	if _, err := os.Stat(targetFile); err == nil {
		Unzip(targetFile, GetCurrentPath())
		OsCommand("chmod +x " + GetCurrentPath() + "*")
		SystemReboot()
	}
}
