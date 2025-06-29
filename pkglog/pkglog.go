package pkglog

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"

	"github.com/Sagiri42/gopkg/pkgutils"
	"gopkg.in/natefinch/lumberjack.v2"
)

func init() {
	New().SetDefaultLog()
}

type Handler struct {
	slog.Handler
	level  *slog.LevelVar
	w      io.Writer
	opts   *slog.HandlerOptions
	format string

	stackTraceSkip     int
	stackTraceMaxDepth int
	mu                 sync.Mutex
}

func New(opts ...HandlerOption) (h *Handler) {
	h = &Handler{
		level:              new(slog.LevelVar),
		w:                  os.Stdout,
		stackTraceSkip:     4,
		stackTraceMaxDepth: 3,
	}
	h.level.Set(slog.LevelInfo)
	h.opts = &slog.HandlerOptions{
		AddSource: true,
		Level:     h.level,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.SourceKey {
				if source, ok := a.Value.Any().(*slog.Source); ok {
					a.Value = slog.StringValue(fmt.Sprintf("%s:%d", source.Function, source.Line))
				}
			}
			return a
		},
	}

	for _, opt := range opts {
		opt(h)
	}

	switch strings.ToUpper(h.format) {
	case "JSON":
		h.Handler = slog.NewJSONHandler(h.w, h.opts)
	case "TEXT":
		fallthrough
	default:
		h.Handler = slog.NewTextHandler(h.w, h.opts)
	}
	return
}

// Handle 重写slog.Handler的Handle方法, 添加错误堆栈信息
func (h *Handler) Handle(ctx context.Context, r slog.Record) (err error) {
	if traceId := getTraceID(ctx); traceId != "" {
		r.AddAttrs(slog.String("traceId", traceId))
	}
	if r.Level == slog.LevelError {
		r.AddAttrs(slog.Any("stack", getStackTrace(h.stackTraceSkip, h.stackTraceMaxDepth)))
	}
	return h.Handler.Handle(ctx, r)
}

// SetLevel 设置日志等级
func (h *Handler) SetLevel(level slog.Level) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.level.Set(level)
}

// SetDefaultLog 设置slog的默认Handler
func (h *Handler) SetDefaultLog() {
	slog.SetDefault(slog.New(h))
}

/*
 * HandlerOption 选项配置
 */

type HandlerOption func(h *Handler)

// HandlerWithIoWriter 修改配置io写入
func HandlerWithIoWriter(w io.Writer) HandlerOption {
	return func(h *Handler) {
		h.w = w
	}
}

// HandlerWithFormat 设置日志输出格式: text, json
func HandlerWithFormat(format string) HandlerOption {
	return func(h *Handler) {
		h.format = format
	}
}

// HandlerWithLevel 设置日志等级
func HandlerWithLevel(level slog.Level) HandlerOption {
	return func(h *Handler) {
		h.level.Set(level)
	}
}

// HandlerWithStrLevel 设置日志等级(字符串)
func HandlerWithStrLevel(level string) HandlerOption {
	return func(h *Handler) {
		switch strings.ToUpper(level) {
		case "DEBUG":
			h.level.Set(slog.LevelDebug)
		case "INFO":
			h.level.Set(slog.LevelInfo)
		case "WARN", "WARNING":
			h.level.Set(slog.LevelWarn)
		case "ERROR":
			h.level.Set(slog.LevelError)
		default:
			slog.Warn("未知的日志等级, 已设置为默认的INFO等级", slog.String("level", level))
			h.level.Set(slog.LevelInfo)
		}
	}
}

// HandlerWithHandlerOptions 设置处理器配置
func HandlerWithHandlerOptions(opts *slog.HandlerOptions) HandlerOption {
	return func(h *Handler) {
		h.opts = opts
	}
}

// HandlerWithStackTrace 设置StackTrace堆栈跟踪信息
func HandlerWithStackTrace(skip, maxDepth int) HandlerOption {
	return func(h *Handler) {
		h.stackTraceSkip = skip
		h.stackTraceMaxDepth = maxDepth
	}
}

// 日志切割配置
func HandlerWithLumberjack(lumberjackLog *lumberjack.Logger) HandlerOption {
	return func(h *Handler) {
		h.w = io.MultiWriter(h.w, lumberjackLog)
	}
}

/*
 * 辅助方法
 */

var TraceId string = "traceId"

// getTraceID 安全获取追踪ID
func getTraceID(ctx context.Context) string {
	if traceID, ok := ctx.Value("traceId").(string); ok {
		return traceID
	}
	return ""
}

// getStackTrace 获得堆栈跟踪信息
func getStackTrace(skip, maxDepth int) []string {
	pcs := make([]uintptr, maxDepth)
	n := runtime.Callers(skip, pcs)
	if n == 0 {
		return nil
	}

	pcs = pcs[:n]
	frames := runtime.CallersFrames(pcs)
	stack := make([]string, 0, n)

	for {
		frame, more := frames.Next()
		stack = append(stack, frame.Function+":"+strconv.Itoa(frame.Line))
		if !more || len(stack) >= maxDepth {
			break
		}
	}
	return stack
}

func DefaultLumberjack() *lumberjack.Logger {
	var (
		err        error
		logDirPath string = filepath.Join(pkgutils.ExecPath(), "logs")
		logDirInfo os.FileInfo
	)
	// 创建日志目录
	if logDirInfo, err = os.Stat(logDirPath); err != nil {
		if err = os.MkdirAll(logDirPath, 0755); err != nil {
			panic(fmt.Sprintf("创建日志目录失败, 日志目录路径: %s", logDirPath))
		}
	} else if !logDirInfo.IsDir() {
		panic(fmt.Sprintf("日志目录路径是一个文件, 日志目录路径: %s", logDirPath))
	}

	return &lumberjack.Logger{
		Filename:   filepath.Join(logDirPath, "app.log"),
		MaxSize:    100,  // MB
		MaxBackups: 10,   // 保留旧日志文件数量
		MaxAge:     30,   // 保留天数
		Compress:   true, // 压缩旧日志
		LocalTime:  true,
	}
}
