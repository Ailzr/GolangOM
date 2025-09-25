package model

import (
	"GolangOM/constant"
	"gorm.io/gorm"
)

type ServerModel struct {
	gorm.Model
	ServerID   string
	IP         string
	Port       int
	User       string
	AuthMethod constant.AuthMethod // password 或 key
	Credential string              // 密钥路径
	Password   string              // 密码 或 密钥的密码
}
