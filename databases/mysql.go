package databases

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	logger "github.com/zztroot/zztlog"
	"zzt_blog/config"
	"zzt_blog/models"
)

var DB *gorm.DB

type MysqlConn struct {
	Config config.ConfigRead
	//DB     *gorm.DB
}

func (m *MysqlConn) ConnectMysql() interface{}{
	return DB
}

func (m *MysqlConn) MysqlOpen() error{
	//读取配置文件
	conf, err := m.Config.RConfigConn()
	if err != nil {
		logger.Error(err)
		return err
	}
	dbHost := conf.Get("mysql.db_host")
	dbPort := conf.Get("mysql.db_port")
	dbUser := conf.Get("mysql.db_user")
	dbPwd := conf.Get("mysql.db_pwd")
	dbName := conf.Get("mysql.db_name")
	conn := fmt.Sprintf(`%s:%s@tcp(%s:%s)/%s?charset=utf8&parseTime=True&loc=Local`, dbUser, dbPwd, dbHost, dbPort, dbName)

	//连接数据库
	DB, err = gorm.Open("mysql", conn)
	if err != nil {
		logger.Error("connect error of database mysql:", err)
		return err
	}
	DB.Set("","ENGINE=InnoDB").AutoMigrate(models.GetObjects()...)
	return err
}



