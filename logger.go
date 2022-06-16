/**
 * @Author: dingqinghui
 * @Description:
 * @File:  Logger
 * @Version: 1.0.0
 * @Date: 2022/5/31 18:07
 */

package activity

import (
	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"io"
	"sync"
)

var (
	log        *logger
	onceLogger sync.Once
)

type LogOption func(*logger)

func WithLogger(log *zap.Logger) LogOption {
	return func(l *logger) {
		l.Logger = log
	}
}

func WithLogConfig(logPath string, logLevel zapcore.Level) LogOption {
	return func(l *logger) {
		l.logPath = logPath
		l.logLevel = logLevel
	}
}

func getLogger() *logger {
	onceLogger.Do(func() {
		log = new(logger)
	})
	return log
}

type logger struct {
	*zap.Logger
	logPath  string
	logLevel zapcore.Level
}

func (m *logger) init(opts ...LogOption) {
	for _, opt := range opts {
		if opt == nil {
			continue
		}
		opt(m)
	}
	if m.Logger == nil {
		m.initLog()
	}
}

func (m *logger) initLog() {
	config := zapcore.EncoderConfig{
		MessageKey:     "M",                                                       // 结构化（json）输出：msg的key
		LevelKey:       "L",                                                       // 结构化（json）输出：日志级别的key（INFO，WARN，ERROR等）
		TimeKey:        "T",                                                       // 结构化（json）输出：时间的key
		CallerKey:      "C",                                                       // 结构化（json）输出：打印日志的文件对应的Key
		NameKey:        "N",                                                       // 结构化（json）输出: 日志名
		StacktraceKey:  "S",                                                       // 结构化（json）输出: 堆栈
		LineEnding:     zapcore.DefaultLineEnding,                                 // 换行符
		EncodeLevel:    zapcore.CapitalLevelEncoder,                               // 将日志级别转换成大写（INFO，WARN，ERROR等）
		EncodeTime:     zapcore.TimeEncoderOfLayout("2006/01/02 15:04:05.000000"), // 日志时间的输出样式
		EncodeDuration: zapcore.SecondsDurationEncoder,                            // 消耗时间的输出样式
		EncodeCaller:   zapcore.ShortCallerEncoder,                                // 采用短文件路径编码输出（test/main.go:14 ）
	}

	loglevel := zap.NewAtomicLevelAt(m.logLevel)
	loggerWriter := m.getLoggerWriter()
	// 实现多个输出
	var cores []zapcore.Core
	// 将info及以下写入logPath，NewConsoleEncoder 是非结构化输出
	cores = append(cores, zapcore.NewCore(zapcore.NewJSONEncoder(config), zapcore.AddSync(loggerWriter), loglevel))
	mulCore := zapcore.NewTee(cores...)
	// 设置初始化字段
	filed := zap.Fields(zap.String("activity", "activity"))
	m.Logger = zap.New(mulCore,
		zap.AddCaller(),
		zap.AddStacktrace(zap.DPanicLevel),
		zap.AddCallerSkip(1),
		filed)
}
func (m *logger) getLoggerWriter() io.Writer {
	var writer = &lumberjack.Logger{
		Filename:   m.logPath,
		MaxSize:    500,  // 最大M数，超过则切割
		MaxBackups: 30,   // 最大文件保留数，超过就删除最老的日志文件
		MaxAge:     30,   // 保存30天
		LocalTime:  true, // 本地时间
		Compress:   true, // 是否压缩
	}
	return writer
}
func (m *logger) getLogger() *zap.Logger {
	return m.Logger
}

func logDebug(msg string, fields ...zap.Field) {
	if log.getLogger() == nil {
		println(msg)
		return
	}
	log.Debug(msg, append(fields, zap.String("operate", ""))...)
}

func logInfo(msg string, fields ...zap.Field) {
	if log.getLogger() == nil {
		println(msg)
		return
	}
	log.Info(msg, append(fields, zap.String("operate", ""))...)
}

func logWarn(msg string, fields ...zap.Field) {
	if log.getLogger() == nil {
		println(msg)
		return
	}
	log.Warn(msg, append(fields, zap.String("operate", ""))...)
}

func logError(msg string, fields ...zap.Field) {
	if log.getLogger() == nil {
		println(msg)
		return
	}
	log.Error(msg, append(fields, zap.String("operate", ""))...)
}

func logDPanic(msg string, fields ...zap.Field) {
	if log.getLogger() == nil {
		println(msg)
		return
	}
	log.DPanic(msg, append(fields, zap.String("operate", ""))...)
}
