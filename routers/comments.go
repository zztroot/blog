package routers

import "zzt_blog/middles"

func (r *RouterStruct) CommentsRouters() {
	//网站留言路由
	c := r.Router.Group("/api/comment")
	c.Use(middles.FilterSql())
	{
		//添加评论
		c.POST("/add", new(middles.CommentHost).CommentVerify(), r.CommentService.UploadComment)
		//查询评论
		c.GET("/query", r.CommentService.QueryComment)
	}
}
