package pkgutils

import (
	"log/slog"
	"os"
	"path/filepath"
)

// ExecPath 获取当前可执行文件的路径
func ExecPath() string {
	execPath, err := os.Executable()
	if err != nil {
		slog.Error("[ExecPath] 获取可执行文件路径失败", "err", err)
		return ""
	}
	return filepath.Dir(execPath)
}

// ExceName 获取当前可执行文件的名称
func ExceName() string {
	execPath, err := os.Executable()
	if err != nil {
		slog.Error("[ExceName] 获取可执行文件路径失败", "err", err)
		return ""
	}
	return filepath.Base(execPath)
}
