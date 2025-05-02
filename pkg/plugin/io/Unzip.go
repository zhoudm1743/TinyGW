package io

import (
	"archive/zip"
	"go.uber.org/zap"
	"io"
	"os"
	"path/filepath"
)

// Unzip 解压缩文件到指定目录中
func Unzip(zipFile, destDir string) error {
	zipReader, err := zip.OpenReader(zipFile)
	if err != nil {
		zap.S().Error("打开文件错误：", err)
		return err
	}
	defer zipReader.Close()

	for _, file := range zipReader.File {
		path := filepath.Join(destDir, file.Name)
		if file.FileInfo().IsDir() {
			os.MkdirAll(path, os.ModePerm)
		} else {
			if err = os.MkdirAll(filepath.Dir(path), os.ModePerm); err != nil {
				zap.S().Error("创建文件夹错误：", err)
				return err
			}

			inFile, err := file.Open()
			if err != nil {
				zap.S().Error("打开文件错误：", err)
				return err
			}
			defer inFile.Close()

			outFile, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
			if err != nil {
				zap.S().Error("打开文件错误：", err)
			}
			defer outFile.Close()

			_, err = io.Copy(outFile, inFile)
			if err != nil {
				zap.S().Error("拷贝文件错误：", err)
			}
		}
	}

	return nil
}
