package crashlog

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime/debug"
	"sync"
	"time"

	"github.com/Sagiri42/gopkg/pkgutils"
)

const (
	timeFormat string = "20060102_150405" // 时间戳格式
)

var (
	once     sync.Once
	logFile  *os.File
	initErr  error
	initLock sync.Mutex
)

// Init 默认初始化: /<可执行文件根目录>/.<可执行文件名>.crashlog
func Init() error {
	return Initialize(pkgutils.ExecPath(), pkgutils.ExceName(), false)
}

// InitWithPath 自定义日志目录: /<自定义日志目录>/.<可执行文件名>.crashlog
func InitWithPath(logDirPath string) error {
	return Initialize(logDirPath, pkgutils.ExceName(), false)
}

// InitToTime 在日志文件名中添加时间格式: /<可执行文件根目录>/.<可执行文件名>.<YYYYmmDD_HHMMSS>.crashlog
func InitToTime() error {
	return Initialize(pkgutils.ExecPath(), pkgutils.ExceName(), true)
}

// InitToTimeWithPath 在日志文件名中添加时间格式: /<自定义日志目录>/.<可执行文件名>.<YYYYmmDD_HHMMSS>.crashlog
func InitToTimeWithPath(logDirPath string) error {
	return Initialize(logDirPath, pkgutils.ExceName(), true)
}

// Initialize
func Initialize(crashLogDirPath, crashLogFileName string, addTimeFormat bool) error {
	initLock.Lock()
	defer initLock.Unlock()

	once.Do(func() {
		var (
			crashLogPath string
		)
		if crashLogDirPath == "" {
			crashLogDirPath = pkgutils.ExecPath()
		}
		if crashLogFileName == "" {
			crashLogFileName = pkgutils.ExceName()
		}
		if addTimeFormat {
			crashLogFileName = fmt.Sprintf("%s.%s", crashLogFileName, time.Now().Format(timeFormat))
		}
		crashLogPath = filepath.Join(crashLogDirPath, fmt.Sprintf(".%s.crashlog", crashLogFileName))
		initErr = initialize(crashLogPath)
	})
	return initErr
}

// initialize 实际初始化逻辑
func initialize(crashLogPath string) (err error) {
	if logFile, err = os.OpenFile(crashLogPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err != nil {
		return fmt.Errorf("创建崩溃日志文件失败(%s): %w", crashLogPath, err)
	}

	if err = debug.SetCrashOutput(logFile, debug.CrashOptions{}); err != nil {
		logFile.Close()
		return fmt.Errorf("设置崩溃输出失败: %w", err)
	}
	return
}

// Close 关闭日志文件（可选调用）
func Close() (err error) {
	initLock.Lock()
	defer initLock.Unlock()

	if logFile != nil {
		err = logFile.Close()
	}
	return nil
}
