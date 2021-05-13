package gzhapi

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/wonderivan/logger"
	"github.com/zztroot/zztlog"
	"zzt_blog/config"
)

type RouterGzh struct {
	gzhRouter *gin.Engine
	Config    config.ConfigRead
	gzh Gzh
}

//新开启一个公众号的gin
func (r *RouterGzh) GzhGin() {
	r.gzhRouter = gin.Default()
	//读取配置文件
	conf, err := r.Config.RConfigConn()
	if err != nil {
		logger.Error(err)
	}
	r.routerGroup()
	ip := conf.GetString("gzhapi.ip")
	port := conf.GetString("gzhapi.port")
	zztlog.InfoF("开启一个新的公众号 Gin服务...[端口号:%s]", port)
	_ = r.gzhRouter.Run(fmt.Sprintf(`%s:%s`, ip, port))
}

func (r *RouterGzh) routerGroup(){
	api := r.gzhRouter.Group("/api/gzh")
	{
		api.GET("/", r.gzh.CheckSignature)
		//测试获取token
		api.GET("/getToken", r.gzh.TestToken)
		//创建菜单
		api.GET("/createMenu", r.gzh.CreateMenu)
	}
}


