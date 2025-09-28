package logs

import (
	"os"

	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var Logger *zap.Logger

func init() {
	devMode := viper.GetBool("Server.Development")
	var err error

	if devMode {
		cfg := zap.NewDevelopmentConfig()
		cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		// in development environment, only output logs to command line, not to log files
		cfg.OutputPaths = []string{"stdout"}
		Logger, err = cfg.Build(zap.AddCaller())
	} else {
		cfg := zap.NewProductionConfig()
		cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		// in production environment, output logs to both command line and log files
		cfg.OutputPaths = []string{"stdout", "logs/app.log"}

		logFile := &lumberjack.Logger{
			Filename:   "logs/app.log",                 // log file path
			MaxSize:    viper.GetInt("Log.MaxSize"),    // single log file size, in MB
			MaxBackups: viper.GetInt("Log.MaxBackups"), // maximum number of log files to retain
			MaxAge:     viper.GetInt("Log.MaxAge"),     // maximum log file retention time, in days
			Compress:   true,                           // whether to compress log files
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
