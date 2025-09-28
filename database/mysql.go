package database

import (
	"GolangOM/logs"
	"fmt"
	"time"

	"github.com/spf13/viper"
	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

func init() {
	var err error
	database := viper.GetString("DB.MySQL.Database")
	host := viper.GetString("DB.MySQL.Host")
	port := viper.GetString("DB.MySQL.Port")
	charset := viper.GetString("DB.MySQL.Charset")
	timezone := viper.GetString("DB.MySQL.TimeZone")
	user := viper.GetString("DB.MySQL.User")
	password := viper.GetString("DB.MySQL.Password")

	arg := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=%s&parseTime=True&loc=%s",
		user,
		password,
		host,
		port,
		database,
		charset,
		timezone,
	)
	connectFailed := true
	// try to connect to database 3 times
	for times := 0; times < 3; times++ {
		DB, err = gorm.Open(mysql.Open(arg), &gorm.Config{})
		if err != nil {
			logs.Logger.Error(fmt.Sprintf("MySQL connect failed, times: %d", times+1), zap.Error(err))
			// wait 1 second after connection failure before retrying
			time.Sleep(1 * time.Second)
			continue
		}
		connectFailed = false
		break
	}
	if connectFailed {
		logs.Logger.Error("MySQL connect failed", zap.Error(err))
	}
	logs.Logger.Info("MySQL connect successfully")
}
