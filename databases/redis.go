package databases

import (
	"fmt"
	"github.com/gomodule/redigo/redis"
	logger "github.com/zztroot/zztlog"
	"zzt_blog/config"
)

type RedisConn struct {
	RedisDB redis.Conn
	Config  config.ConfigRead
}

func (r *RedisConn) ConnectRedis() redis.Conn {
	conf, err := r.Config.RConfigConn()
	if err != nil {
		logger.Error(err)
	}
	ip := conf.Get("redis.db_host")
	port := conf.Get("redis.db_port")
	pwd := conf.GetString("redis.pwd")
	//redis.DialPassword(pwd)
	r.RedisDB, err = redis.Dial("tcp", fmt.Sprintf(`%s:%s`, ip, port))
	if err != nil {
		logger.Error("connect error of databases redis:", err)
		return nil
	}
	_, _ = r.RedisDB.Do("auth", pwd)
	return r.RedisDB
}
