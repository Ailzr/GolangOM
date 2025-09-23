package logs

import (
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"os"
)

var Logger *zap.Logger

func init() {
	devMode := viper.GetBool("Server.Development")
	var err error

	if devMode {
		cfg := zap.NewDevelopmentConfig()
		cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		// 开发环境下仅把日志输出到命令行，并不输出日志文件
		cfg.OutputPaths = []string{"stdout"}
		Logger, err = cfg.Build(zap.AddCaller())
	} else {
		cfg := zap.NewProductionConfig()
		cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		// 生产环境下同时把日志输出到命令行和日志文件
		cfg.OutputPaths = []string{"stdout", "logs/app.log"}

		logFile := &lumberjack.Logger{
			Filename:   "logs/app.log",                 // 日志文件路径
			MaxSize:    viper.GetInt("Log.MaxSize"),    // 单个日志文件大小，单位MB
			MaxBackups: viper.GetInt("Log.MaxBackups"), // 最多保留日志文件数量
			MaxAge:     viper.GetInt("Log.MaxAge"),     // 最长保留日志文件时间，单位天
			Compress:   true,                           // 是否压缩日志文件
		}

		consoleCore := zapcore.NewCore(
			zapcore.NewJSONEncoder(cfg.EncoderConfig),
			zapcore.AddSync(os.Stdout),
			cfg.Level,
		)

		fileCore := zapcore.NewCore(
			zapcore.NewJSONEncoder(cfg.EncoderConfig),
			zapcore.AddSync(logFile),
			cfg.Level,
		)

		core := zapcore.NewTee(consoleCore, fileCore)

		Logger = zap.New(core, zap.AddCaller(), zap.AddStacktrace(zap.ErrorLevel))
	}

	if err != nil {
		panic(err)
	}
}
