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
		zap.S().Error("打开zip文件错误：", err)
		return err
	}
	defer zipReader.Close()

	for _, file := range zipReader.File {
		path := filepath.Join(destDir, file.Name)
		if file.FileInfo().IsDir() {
			// 创建目录并设置正确的权限
			if err := os.MkdirAll(path, file.Mode()); err != nil {
				zap.S().Error("创建目录错误：", err)
				return err
			}
			continue
		}

		// 确保父目录存在
		if err := os.MkdirAll(filepath.Dir(path), os.ModePerm); err != nil {
			zap.S().Error("创建父目录错误：", err)
			return err
		}

		// 打开zip内的文件
		inFile, err := file.Open()
		if err != nil {
			zap.S().Error("打开压缩文件错误：", err)
			return err
		}
		defer inFile.Close()
		// 删除目标文件
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			zap.S().Error("删除目标文件错误：", err)
			return err
		}
		// 创建目标文件
		outFile, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
		if err != nil {
			zap.S().Error("创建目标文件错误：", err)
			return err
		}
		defer outFile.Close()

		// 拷贝文件内容
		if _, err := io.Copy(outFile, inFile); err != nil {
			zap.S().Error("拷贝文件内容错误：", err)
			return err // 关键修复：错误时立即返回
		}
	}

	return nil
}
