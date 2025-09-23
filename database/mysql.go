package database

import (
	"GolangOM/logs"
	"fmt"
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
	DB, err = gorm.Open(mysql.Open(viper.GetString(arg)), &gorm.Config{})
	if err != nil {
		logs.Logger.Fatal("MySQL connect failed", zap.Error(err))
	}
	logs.Logger.Info("MySQL connect successfully")
}
