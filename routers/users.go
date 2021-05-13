package routers

import "zzt_blog/middles"

func (r *RouterStruct) UserRouters(){
	//用户相关路由
	user := r.Router.Group("/api/user")

	//中间件
	user.Use(middles.FilterSql())

	{
		//用户登录
		user.POST("/login", r.UserService.Login)
	}
	{
		//用户信息处理
		user.GET("/query", r.UserService.QueryInfo)
		user.POST("/upload",middles.TokenCheck(), r.UserService.UploadInfo)
	}
	{
		//第三方账号登录回调
		//github
		user.GET("/githubCallback", r.UserService.GithubCallback)
	}
}
