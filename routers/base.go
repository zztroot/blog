package routers

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"os"
	"path"
	"time"
	gzhapi "zzt_blog/gzhApi"

	//"github.com/wonderivan/logger"
	logger "github.com/zztroot/zztlog"
	"zzt_blog/config"
	"zzt_blog/databases"
	"zzt_blog/middles"
	"zzt_blog/mvc"
	"zzt_blog/services"
)

type RouterStruct struct {
	Config         config.ConfigRead
	ArticleService services.ArticleService
	UserService    services.UserService
	CommentService services.CommentService
	ToolsService   services.ToolsService
	Router         *gin.Engine
	mvc.RouterMvc
	gzhapi.RouterGzh
}

const (
	MediaPath = "media"
)

func (r *RouterStruct) Init() {
	//通过文件配置log
	if err := logger.InitConfigFile("log.json"); err != nil {
		logger.Error(err)
		return
	}
	//开启一个新的gin服务做MVC开发
	go r.MvcGin()

	//开启一个新的gin API做公众号开发
	go r.GzhGin()

	//数据清理
	go clearFile()

	//读取配置文件
	conf, err := r.Config.RConfigConn()
	if err != nil {
		logger.Error(err)
	}
	//连接数据库
	mysql := databases.MysqlConn{}
	err = mysql.MysqlOpen()
	if err != nil {
		logger.Error(err)
		return
	}
	db := mysql.ConnectMysql()

	ip := conf.Get("server.ip")
	port := conf.Get("server.port")
	apiMode := conf.Get("server.run")
	r.Router = gin.Default()
	if apiMode != "debug" {
		gin.SetMode(gin.ReleaseMode)
	}
	r.Router.Use(middles.Cors())
	//向百度提交url
	go middles.PushBaiduUrl(db)
	//文章路由
	r.ArticleRouters()
	//账号路由
	r.UserRouters()
	//网站留言
	r.CommentsRouters()
	//工具集
	r.ToolsRouters()

	//r.FileRouters()
	_ = r.Router.Run(fmt.Sprintf(`%s:%s`, ip, port))
}

//清理media文件
func clearFile() {
	logger.Info("进入数据删除")
	for {
		fs, _ := ioutil.ReadDir(MediaPath)
		if len(fs) <= 1 {
			return
		}
		for _, v := range fs {
			timeLimit := time.Now().AddDate(0, 0, -5).Format(`2006-01-02 15:04:05`)
			fileTime := v.ModTime().Format(`2006-01-02 15:04:05`)

			if fileTime < timeLimit {
				if err := os.RemoveAll(path.Join(MediaPath, v.Name())); err != nil {
					logger.Error(err)
				} else {
					logger.InfoF("成功删除%s文件", v.Name())
				}
			}
		}
		//每24小时清理一次
		time.Sleep(time.Hour * 24)
	}
}
