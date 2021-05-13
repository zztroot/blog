package mvc

import (
	"fmt"
	"github.com/gin-gonic/gin"
	logger "github.com/zztroot/zztlog"
	"zzt_blog/config"
	"zzt_blog/mvc/handlers"
)

type RouterMvc struct {
	mvcRouter *gin.Engine
	Config    config.ConfigRead
	articleHandler handlers.ArticleHandlerMvc
}

//新开启一个mvc的gin
func (r *RouterMvc) MvcGin() {
	r.mvcRouter = gin.Default()
	r.mvcRouter.LoadHTMLGlob("./templates/**/*")
	r.mvcRouter.Static("/static", "./static")

	//读取配置文件
	conf, err := r.Config.RConfigConn()
	if err != nil {
		logger.Error(err)
	}
	r.MvcArticleRouter()
	ip := conf.GetString("mvcserver.ip")
	port := conf.GetString("mvcserver.port")
	logger.InfoF("开启一个新的MVC Gin服务...[端口号:%s]", port)
	_ = r.mvcRouter.Run(fmt.Sprintf(`%s:%s`, ip, port))
}

//获取单个文章(单页面MVC)
func (r *RouterMvc) MvcArticleRouter() {
	//文章相关路由
	article := r.mvcRouter.Group("/mvc")
	{
		article.GET("/article/get", r.articleHandler.ArticleGetMvc)
	}
}
