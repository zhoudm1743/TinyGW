package io

import (
	"fmt"
	"go.uber.org/zap"
	"io"
	"io/ioutil"
	"net/url"
	"os"
	"path"
	"strings"
)

func GetLogPath() string {
	return path.Join(GetCurrentPath(), "public", "logs", "energy.log")
}

func GetCurrentPath() string {
	cur, err := os.Getwd()
	if err != nil {
		zap.S().Errorf("获取当前路径失败 %v", err)
	}
	return cur + "/"
}

func GetPluginPath() string {
	return GetCurrentPath() + "plugin/"
}

func SureExists(dir string) {
	cur := GetCurrentPath()
	_, err := os.Stat(cur + dir)
	if err != nil {
		err := os.Mkdir(cur+dir, 0777)
		if err != nil {
			zap.S().Errorf("创建%s文件夹失败 %v", cur+dir, err)
		}
		_ = os.Chmod(cur+dir, 0777)
	}
}

func PathExists(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		return os.IsExist(err)
	}
	return true
}

func ReadFile(filename string) ([]byte, error) {
	return os.ReadFile(GetCurrentPath() + filename)
}

func WriteFile(filename string, data []byte) error {
	return os.WriteFile(GetCurrentPath()+filename, data, 0777)
}

// File copies a single file from src to dst
func File(src, dst string) error {
	zap.S().Debug(src)
	zap.S().Debug(dst)
	var err error
	var srcfd *os.File
	var dstfd *os.File
	var srcinfo os.FileInfo

	if srcfd, err = os.Open(src); err != nil {
		return err
	}
	defer srcfd.Close()

	if dstfd, err = os.Create(dst); err != nil {
		return err
	}
	defer dstfd.Close()
	zap.S().Debug(dstfd, srcfd)
	if _, err = io.Copy(dstfd, srcfd); err != nil {
		return err
	}
	if srcinfo, err = os.Stat(src); err != nil {
		return err
	}
	return os.Chmod(dst, srcinfo.Mode())
}

// Dir copies a whole directory recursively
func Dir(src string, dst string) error {
	var err error
	var fds []os.FileInfo
	var srcinfo os.FileInfo

	if srcinfo, err = os.Stat(src); err != nil {
		return err
	}

	if err = os.MkdirAll(dst, srcinfo.Mode()); err != nil {
		return err
	}

	if fds, err = ioutil.ReadDir(src); err != nil {
		return err
	}
	for _, fd := range fds {
		srcfp := path.Join(src, fd.Name())
		dstfp := path.Join(dst, fd.Name())

		if fd.IsDir() {
			if err = Dir(srcfp, dstfp); err != nil {
				fmt.Println(err)
			}
		} else {
			if err = File(srcfp, dstfp); err != nil {
				fmt.Println(err)
			}
		}
	}
	return nil
}

func DownloadFile(sourceFile, targetFile string) {
	remoteFile, _ := url.Parse(sourceFile)
	OsCommand("wget -O " + targetFile + " " + remoteFile.String())
}

func ReadLastNLines(path string, n int) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	stat, _ := file.Stat()
	size := stat.Size()
	buf := make([]byte, 512)
	lineCount := 0
	pos := size

	for lineCount < n && pos > 0 {
		readSize := int64(len(buf))
		if pos < readSize {
			readSize = pos
		}
		pos -= readSize

		_, err = file.Seek(pos, io.SeekStart)
		if err != nil {
			return nil, err
		}

		if _, err = io.ReadFull(file, buf[:readSize]); err != nil && err != io.EOF {
			return nil, err
		}

		for i := readSize - 1; i >= 0; i-- {
			if buf[i] == '\n' {
				lineCount++
				if lineCount >= n {
					pos += i + 1
					break
				}
			}
		}
	}

	_, err = file.Seek(pos, io.SeekStart)
	if err != nil {
		return nil, err
	}

	content, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(strings.TrimSuffix(string(content), "\n"), "\n")
	if len(lines) > n {
		lines = lines[len(lines)-n:]
	}

	return lines, nil
}
