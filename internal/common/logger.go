/*
 * @Date: 2026-04-16 10:00:43
 * @Author: zhongwenhao
 * @LastEditors: zhongwenhao
 * @LastEditTime: 2026-04-17 14:00:42
 * @Description: 日志记录器	由于zap不具备日志切割功能, 这里使用lumberjack配合使用
 */

package common

import (
	"fmt"
	"go-blog/internal/config"
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// Sugared Logger 并重性能与易用性，支持结构化和 printf 风格的日志记录。
// Logger 非常强调性能，不提供 printf 风格的 api （减少了 interface{} 与 反射的性能损耗）
var Log *zap.SugaredLogger

/**
 * 初始化日志
 * filename 日志文件路径
 * level 日志级别
 * maxSize 每个日志文件保存的最大尺寸 单位：M
 * maxBackups 日志文件最多保存多少个备份
 * maxAge 文件最多保存多少天
 * compress 是否压缩
 * serviceName 服务名
 */
func InitLogger() {
	var coreArr []zapcore.Core

	// 1. 定义不同输出目标的级别过滤器

	// Zap 提供了 Debug, Info, Warn, Error, DPanic, Panic, Fatal 七个日志级别，优先级逐级递增
	// 仅记录 Error 级别及以上的日志
	highPriority := zap.LevelEnablerFunc(func(level zapcore.Level) bool {
		return level >= zap.ErrorLevel
	})
	// 仅记录 Info 级别及以上的日志
	lowPriority := zap.LevelEnablerFunc(func(level zapcore.Level) bool {
		return level < zap.ErrorLevel && level >= zap.DebugLevel
	})
	// 当yml配置中的等级大于Error时，lowPriority级别日志停止记录
	if config.Conf.Logs.Level >= 2 {
		lowPriority = zap.LevelEnablerFunc(func(level zapcore.Level) bool {
			return false
		})
	}

	// 2. 配置日志编码器
	encoder := getEncoder()

	// 3. 配置日志输出目标 (WriteSyncer)
	infoFileWriteSyncer := getInfoWriteSyncer()
	errorFileWriteSyncer := getErrorWriteSyncer()

	// 4. 创建多个 Core
	infoFileCore := zapcore.NewCore(encoder, zapcore.NewMultiWriteSyncer(infoFileWriteSyncer, zapcore.AddSync(os.Stdout)), lowPriority)
	errorFileCore := zapcore.NewCore(encoder, zapcore.NewMultiWriteSyncer(errorFileWriteSyncer, zapcore.AddSync(os.Stdout)), highPriority)

	coreArr = append(coreArr, infoFileCore)
	coreArr = append(coreArr, errorFileCore)
	// 5. 创建 Logger，使用 NewTee 将多个 Core 合并，并添加调用者信息(行号、文件名)
	logger := zap.New(zapcore.NewTee(coreArr...), zap.AddCaller())
	// 6. 创建 Sugared Logger，方便使用
	Log = logger.Sugar()
	Log.Info("初始化zap日志完成!")
}

/**
 * 获取编码器
 */
func getEncoder() zapcore.Encoder {
	encoderConfig := zapcore.EncoderConfig{
		MessageKey:    "msg",
		LevelKey:      "level",
		TimeKey:       "time",
		NameKey:       "name",
		CallerKey:     "file",
		FunctionKey:   "func",
		StacktraceKey: "stacktrace",
		LineEnding:    zapcore.DefaultLineEnding,
		EncodeLevel:   zapcore.CapitalLevelEncoder,
		EncodeTime: func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
			enc.AppendString(t.Format("2006-01-02 15:04:05"))
		},
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}
	// 配置日志编码器 (生产环境推荐使用 JSON)
	encoder := zapcore.NewConsoleEncoder(encoderConfig)
	return encoder
}

/**
 * 获取信息文件写入器
 */
func getInfoWriteSyncer() zapcore.WriteSyncer {
	now := time.Now()
	infoLogFileName := fmt.Sprintf("%s/info/%04d-%02d-%02d.log", config.Conf.Logs.Path, now.Year(), now.Month(), now.Day())
	return zapcore.AddSync(&lumberjack.Logger{
		Filename:   infoLogFileName,               //日志文件存放目录，如果文件夹不存在会自动创建
		MaxSize:    config.Conf.Logs.MaxSize,    //文件大小限制,单位MB
		MaxAge:     config.Conf.Logs.MaxAge,     //日志文件保留天数
		MaxBackups: config.Conf.Logs.MaxBackups, //最大保留日志文件数量
		LocalTime:  false,
		Compress:   config.Conf.Logs.Compress, //是否压缩处理
	})
}

/**
 * 获取错误文件写入器
 */
func getErrorWriteSyncer() zapcore.WriteSyncer {
	now := time.Now()
	errorLogFileName := fmt.Sprintf("%s/error/%04d-%02d-%02d.log", config.Conf.Logs.Path, now.Year(), now.Month(), now.Day())
	return zapcore.AddSync(&lumberjack.Logger{
		Filename:   errorLogFileName,              //日志文件存放目录
		MaxSize:    config.Conf.Logs.MaxSize,    //文件大小限制,单位MB
		MaxAge:     config.Conf.Logs.MaxAge,     //日志文件保留天数
		MaxBackups: config.Conf.Logs.MaxBackups, //最大保留日志文件数量
		LocalTime:  false,
		Compress:   config.Conf.Logs.Compress, //是否压缩处理
	})
}
