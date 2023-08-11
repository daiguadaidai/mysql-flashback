package utils

import (
	"fmt"
	"github.com/cihub/seelog"
	"os"
	"os/exec"
	"path/filepath"
)

// 文件/目录 是否存在
func PathExists(p string) (bool, error) {
	_, err := os.Stat(p)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}

	return false, err
}

// 创建一个目录
func CreateDir(p string) error {
	err := os.MkdirAll(p, os.ModePerm)

	return err
}

// 检测和创建目录是否存在
func CheckAndCreatePath(p string, msg string) error {
	exists, err := PathExists(p)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}

	// 路径不存在则创建
	seelog.Warnf("%s不存在(创建中): %s", msg, p)
	if err = CreateDir(p); err != nil {
		return fmt.Errorf("%s(创建失败): %s. %v", msg, p, err)
	}
	seelog.Warnf("%s(创建成功): %s", msg, p)

	return nil
}

// 获取文件绝对路径
func FileAbs(filePath string) (string, error) {
	return filepath.Abs(filePath)
}

func Filename(filePath string) string {
	return filepath.Base(filePath)
}

// 获取执行文件的路径
func CMDDir() (string, error) {
	filePath, err := exec.LookPath(os.Args[0])
	if err != nil {
		return "", err
	}

	fileAbs, err := FileAbs(filePath)
	if err != nil {
		return "", err
	}

	rst := filepath.Dir(fileAbs)
	return rst, nil
}
