package config

import (
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"os"
)

func init() {

	logger, _ := zap.NewDevelopment()
	defer logger.Sync()
	// 设置配置文件路径和类型
	workDir, _ := os.Getwd()
	viper.SetConfigName("configs.yaml")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(workDir + "/config")

	if err := viper.ReadInConfig(); err != nil {
		logger.Fatal("load config file failed", zap.Error(err))
	}
	logger.Info("load config file successfully")

}
