package util

import "crypto/md5"

func GetMd5(str string) string {
	return string(md5.New().Sum([]byte(str)))
}
