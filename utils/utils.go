package utils

import (
	"crypto/md5"
	"encoding/hex"
	"github.com/spf13/cast"
	"github.com/zheng-ji/goSnowFlake"
)

var iw *goSnowFlake.IdWorker

func init() {
	iw, _ = goSnowFlake.NewIdWorker(1)
}
func GetSnowId() string {
	id, _ := iw.NextId()
	return Md5Str(cast.ToString(id))
}

func Md5Str(s string) string {
	hash := md5.New()
	hash.Write([]byte(s))
	value := hash.Sum(nil)
	return hex.EncodeToString(value)
}

