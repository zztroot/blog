package handlers

import "github.com/gin-gonic/gin"

type ArticleHandlerMvc struct {

}

func (a *ArticleHandlerMvc) ArticleGetMvc (c *gin.Context){
	c.HTML(200, "article.html", nil)
}
