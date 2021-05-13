package routers

import "zzt_blog/middles"

func (r *RouterStruct) ArticleRouters() {
	//文章相关路由
	article := r.Router.Group("/api/article")
	//中间件
	article.Use(middles.FilterSql())
	{ //文章
		//文章上传
		article.POST("/upload", middles.TokenCheck(), r.ArticleService.ArticleUpload)
		//文章删除
		article.POST("/del", middles.TokenCheck(), r.ArticleService.ArticleDel)
		//文章获取(分页)
		article.GET("/getAll/:line/:page", r.ArticleService.ArticleGetAll)
		//文章获取(不分页)
		article.GET("/getAllNot", r.ArticleService.ArticleGetAllNotPage)
		//文章编辑
		article.POST("/modify", middles.TokenCheck(), r.ArticleService.ArticleModify)
		//文章获取单个
		article.GET("/get/:id", r.ArticleService.ArticleGet)
		//文章评论添加
		article.POST("/add/comment", r.ArticleService.ArticleCommentAdd)
		//文章评论查询
		article.POST("/query/comment", r.ArticleService.ArticleCommentQuery)
		//查询最火的文章
		article.GET("/query/hot", r.ArticleService.ArticleHot)
		//查询最新的文章
		article.GET("/query/latest", r.ArticleService.ArticleLatest)
		//搜索文章
		article.POST("/search", r.ArticleService.ArticleSearch)
	}
	{ //类型
		//类型上传
		article.POST("/type/upload", middles.TokenCheck(), r.ArticleService.ArticleTypeUpload)
		//类型查询
		article.GET("/type/getAll", r.ArticleService.ArticleTypeGetAll)
		//类型删除
		article.POST("/type/del", middles.TokenCheck(), r.ArticleService.ArticleTypeDel)
	}
}
