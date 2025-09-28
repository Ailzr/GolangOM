package config

import (
	"os"

	"github.com/spf13/viper"
	"go.uber.org/zap"
)

func init() {

	logger, _ := zap.NewDevelopment()
	defer logger.Sync()
	// set configuration file path and type
	workDir, _ := os.Getwd()
	viper.SetConfigName("configs.yaml")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(workDir + "/config")

	if err := viper.ReadInConfig(); err != nil {
		logger.Fatal("load config file failed", zap.Error(err))
	}
	logger.Info("load config file successfully")

}
