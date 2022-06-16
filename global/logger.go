/**
 * @Author: dingqinghui
 * @Description:
 * @File:  Logger
 * @Version: 1.0.0
 * @Date: 2022/5/31 18:07
 */

package global

import "go.uber.org/zap"

var (
	Logger *zap.Logger
	tag    = "[operate]"
)

func LogDebug(msg string, fields ...zap.Field) {
	if Logger != nil {
		Logger.Debug(msg, append(fields, zap.String("operate", ""))...)
	} else {
		println(tag, "[debug]", msg)
	}
}
func LogInfo(msg string, fields ...zap.Field) {
	if Logger != nil {
		Logger.Info(msg, append(fields, zap.String("operate", ""))...)
	} else {
		println(tag, "[info]", msg)
	}
}
func LogWarn(msg string, fields ...zap.Field) {
	if Logger != nil {
		Logger.Warn(msg, append(fields, zap.String("operate", ""))...)
	} else {
		println(tag, "[warn]", msg)
	}
}

func LogError(msg string, fields ...zap.Field) {
	if Logger != nil {
		Logger.Error(msg, append(fields, zap.String("operate", ""))...)
	} else {
		println(tag, "[error]", msg)
	}
}

func LogDPanic(msg string, fields ...zap.Field) {
	if Logger != nil {
		Logger.DPanic(msg, append(fields, zap.String("operate", ""))...)
	} else {
		println(tag, "[panic]", msg)
	}
}
