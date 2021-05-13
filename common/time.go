package common

import "time"

//获取当前时间
func GetTime() string {
	return time.Now().Format("2006年01月02日 15:04:05")
}

func GetTimeGzh() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

