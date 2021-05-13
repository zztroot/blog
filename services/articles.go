package services

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gomodule/redigo/redis"
	"github.com/jinzhu/gorm"
	"github.com/levigross/grequests"
	logger "github.com/zztroot/zztlog"
	"strconv"
	. "zzt_blog/common"
	"zzt_blog/databases"
	"zzt_blog/models"
)

type ArticleService struct {
	db  databases.MysqlConn
	rdb databases.RedisConn
	mdb databases.MongoConn
}

//文章搜索
func (a *ArticleService) ArticleSearch(c *gin.Context) {
	temp, _ := c.Get("data")
	data := temp.(map[string]interface{})
	db := a.db.ConnectMysql().(*gorm.DB)

	searchStr := data["search"].(string)

	var articles []models.Article
	if err := db.Debug().Where("article_name LIKE ? and delete_at = ?", fmt.Sprintf("%%%s%%", searchStr), "no").Find(&articles).Error; err != nil {
		logger.Error(err)
		return
	}
	m := make(map[string]interface{})
	m["data"] = articles
	bs, _ := json.Marshal(m)
	Reply(c, SlotReply{Data: bs})
}

//查询最新5篇文章
func (a *ArticleService) ArticleLatest(c *gin.Context) {
	db := a.db.ConnectMysql().(*gorm.DB)
	var articles []models.Article
	if err := db.Where("delete_at = ?", "no").Order("created_at desc").Limit(5).Find(&articles).Error; err != nil {
		logger.Error(err)
		return
	}
	//var temp []models.Article
	//for _, v := range articles {
	//	tempRune := []rune(v.ArticleName)
	//	if len(v.ArticleName) > 30 {
	//		v.ArticleName = string(tempRune[:29])
	//		temp = append(temp, v)
	//	} else {
	//		temp = append(temp, v)
	//	}
	//}
	m := make(map[string]interface{})
	m["data"] = articles
	bs, _ := json.Marshal(m)
	Reply(c, SlotReply{Data: bs})
}

//查询阅读量最多的5篇文章
func (a *ArticleService) ArticleHot(c *gin.Context) {
	db := a.db.ConnectMysql().(*gorm.DB)
	rdb := a.rdb.ConnectRedis()

	//查询
	var articles []models.Article
	if err := db.Where("delete_at = ?", "no").Find(&articles).Error; err != nil {
		logger.Error(err)
		return
	}
	var temp []models.Article
	for _, v := range articles {
		readCount, _ := redis.Int(rdb.Do("get", v.Id))
		v.ReadCount = uint(readCount)
		temp = append(temp, v)
	}
	for i, _ := range temp {
		for x, _ := range temp {
			if temp[i].ReadCount < temp[x].ReadCount {
				continue
			} else {
				t := temp[x]
				temp[x] = temp[i]
				temp[i] = t
			}
		}
	}

	//处理标题长度
	//var tempArticles []models.Article
	//for _, v := range temp[len(temp)-5 : len(temp)] {
	//	tempRune := []rune(v.ArticleName)
	//	if len(v.ArticleName) > 30 {
	//		v.ArticleName = string(tempRune[:29])
	//		tempArticles = append(tempArticles, v)
	//	} else {
	//		tempArticles = append(tempArticles, v)
	//	}
	//}

	m := make(map[string]interface{})
	m["data"] = temp[:5]
	bs, _ := json.Marshal(m)
	Reply(c, SlotReply{Data: bs})
}

//文章评论查询(all)
func (a *ArticleService) ArticleCommentQuery(c *gin.Context) {
	temp, _ := c.Get("data")
	data := temp.(map[string]interface{})
	db := a.db.ConnectMysql().(*gorm.DB)

	//查询
	var comments []models.ArticleComment
	if err := db.Where("article_id = ?", data["article_id"].(float64)).Order("created_at desc").Find(&comments).Error; err != nil {
		logger.Error(err)
		return
	}

	//返回结果
	m := make(map[string]interface{})
	m["data"] = comments
	bs, _ := json.Marshal(m)
	Reply(c, SlotReply{Data: bs})
}

//文章评论添加
func (a *ArticleService) ArticleCommentAdd(c *gin.Context) {
	temp, _ := c.Get("data")
	data := temp.(map[string]interface{})
	db := a.db.ConnectMysql().(*gorm.DB)

	//判断评论内容是否为空
	if data["user_name"].(string) == "" || data["content"].(string) == "" {
		Reply(c, SlotReply{Error: "数据不能为空"})
		return
	}

	comment := models.ArticleComment{}
	comment.ArticleId = uint(data["article_id"].(float64))
	comment.TempUserName = data["user_name"].(string)
	comment.Content = data["content"].(string)
	comment.CreatedAt = GetTime()

	msg := "评论成功"
	if data["to_user_name"] != nil {
		comment.ToTempUserName = data["to_user_name"].(string)
		msg = "回复成功"
	}
	if err := db.Save(&comment).Error; err != nil {
		logger.Error(err)
		return
	}
	m := make(map[string]interface{})
	m["data"] = msg
	ms, _ := json.Marshal(m)
	Reply(c, SlotReply{Data: ms})
}

//文章编辑
func (a *ArticleService) ArticleModify(c *gin.Context) {
	temp, _ := c.Get("data")
	data := temp.(map[string]interface{})
	db := a.db.ConnectMysql().(*gorm.DB)
	//判断
	if data["article_name"].(string) == "" || data["article_content"].(string) == "" {
		logger.Error("请输入内容")
		Reply(c, SlotReply{Error: "请输入内容"})
		return
	}
	//查询类型ID是否存在
	types := models.Type{}
	if err := db.Where("type_name =?", data["types_name"].(string)).Find(&types).Error; err != nil {
		logger.Error("没有此类型")
		return
	}
	if data["id"] == nil {
		logger.Error("缺少参数")
		return
	}
	ids := data["id"].(map[string]interface{})
	id := ids["_value"].(string)
	article := models.Article{}
	t, _ := strconv.Atoi(id)
	article.Id = uint(t)
	if err := db.Find(&article).Error; err != nil {
		logger.Error(err)
		return
	}
	var tempArticle = models.Article{}
	tempArticle.Id = article.Id
	tempArticle.ArticleName = data["article_name"].(string)
	tempArticle.ArticleContent = data["article_content"].(string)
	tempArticle.TypesId = types.Id
	tempArticle.CreatedAt = article.CreatedAt
	tempArticle.DeleteAt = article.DeleteAt
	tempArticle.UpdatedAt = GetTime()
	if err := db.Save(&tempArticle).Error; err != nil {
		logger.Error(err)
		return
	}

	m := make(map[string]interface{})
	m["data"] = "修改成功"
	bs, _ := json.Marshal(m)
	Reply(c, SlotReply{Data: bs})
}

//文章删除
func (a *ArticleService) ArticleDel(c *gin.Context) {
	temp, _ := c.Get("data")
	data := temp.(map[string]interface{})
	db := a.db.ConnectMysql().(*gorm.DB)

	if data["id"] == nil {
		logger.Error("缺少参数")
		return
	}
	id := data["id"].(float64)
	article := models.Article{}
	article.Id = uint(id)
	if err := db.Model(&article).Update("delete_at", "yes").Error; err != nil {
		logger.Error(err)
		return
	}

	m := make(map[string]interface{})
	m["data"] = "删除成功"
	bs, _ := json.Marshal(m)
	Reply(c, SlotReply{Data: bs})
}

//文章上传
func (a *ArticleService) ArticleUpload(c *gin.Context) {
	data, _ := c.Get("data")
	tempData := data.(map[string]interface{})

	//判断
	if tempData["article_name"].(string) == "" || tempData["article_content"].(string) == "" {
		logger.Error("请输入内容")
		Reply(c, SlotReply{Error: "请输入内容"})
		return
	}

	//获取mysql连接对象
	db := a.db.ConnectMysql().(*gorm.DB)

	//查询类型ID是否存在
	types := models.Type{}
	if err := db.Where("type_name =?", tempData["types_name"].(string)).Find(&types).Error; err != nil {
		logger.Error("没有此类型")
		return
	}

	//保存文章
	article := models.Article{}
	article.ArticleName = tempData["article_name"].(string)
	article.ArticleContent = tempData["article_content"].(string)
	article.TypesId = types.Id
	article.CreatedAt = GetTime()
	article.DeleteAt = "no"
	if err := db.Save(&article).Error; err != nil {
		logger.Error(err)
		return
	}

	//创建文章的redis(阅读量)
	rdb := a.rdb.ConnectRedis()
	//_, _ = rdb.Do("auth", "123456789")
	_, err := rdb.Do("SET", strconv.Itoa(int(article.Id)), 0)
	if err != nil {
		logger.Error(err)
		return
	}

	//将文章推送给百度收录
	url := "http://data.zz.baidu.com/urls?site=https://www.zztdd.cn&token=uvqkZC9WzyCGL1lJ"
	re, err := grequests.Post(url, &grequests.RequestOptions{
		Headers: map[string]string{"Content-Type": "text/plain"},
		JSON:    fmt.Sprintf(`https://www.zztdd.cn/showArticle?id=%d`, article.Id),
	})
	if err != nil {
		logger.Error(err)
		return
	}
	logger.Info(string(re.Bytes()))

	m := make(map[string]interface{})
	m["data"] = "上传成功"
	bs, _ := json.Marshal(m)
	Reply(c, SlotReply{Data: bs})
}

//获取所有文章(不分页)
func (a *ArticleService) ArticleGetAllNotPage(c *gin.Context) {
	db := a.db.ConnectMysql().(*gorm.DB)
	rdb := a.rdb.ConnectRedis()
	var articles []models.Article
	if err := db.Where("delete_at =?", "no").Order("created_at desc").Find(&articles).Error; err != nil {
		logger.Error(err)
		return
	}

	//查询出文章的阅读数
	for _, v := range articles {
		readCount, _ := redis.Int(rdb.Do("get", v.Id))
		v.ReadCount = uint(readCount)
	}
	m := make(map[string]interface{})
	m["total"] = len(articles)
	m["data"] = articles
	bs, _ := json.Marshal(m)
	Reply(c, SlotReply{Data: bs})
}

//获取所有文章(分页)
func (a *ArticleService) ArticleGetAll(c *gin.Context) {
	db := a.db.ConnectMysql().(*gorm.DB)
	rdb := a.rdb.ConnectRedis()

	//分页处理
	line := c.Param("line")
	tempLine, _ := strconv.Atoi(line)
	page := c.Param("page")
	tempPage, _ := strconv.Atoi(page)
	newPage := (tempPage - 1) * tempLine
	var articles []models.Article

	//分页查询
	if err := db.Where("delete_at = ?", "no").Order("created_at desc").Limit(line).Offset(newPage).Find(&articles).Error; err != nil {
		logger.Error(err)
		return
	}

	var article []models.Article
	var count int
	if err := db.Where("delete_at = ?", "no").Find(&article).Count(&count).Error; err != nil {
		logger.Error(err)
		return
	}

	type articleResults struct {
		Id           uint   `gorm:"primary_key" json:"id"`
		CreatedAt    string `json:"created_time"`
		ArticleName  string `json:"article_name"`
		TypesId      uint   `json:"types_id"`
		ReadCount    uint   `json:"read_count"`
		LikeCount    uint   `json:"like_count"`
		CommentCount uint   `json:"comment_count"`
	}

	var temp []articleResults
	//查询每篇文章的阅读量
	for _, v := range articles {
		var t articleResults
		readCount, err := redis.Int(rdb.Do("get", v.Id))
		if err != nil {
			logger.Error(err)
		}
		t.Id = v.Id
		t.ArticleName = v.ArticleName
		t.TypesId = v.TypesId
		t.ReadCount = uint(readCount)
		t.LikeCount = v.LikeCount
		t.CommentCount = v.CommentCount
		t.CreatedAt = v.CreatedAt
		temp = append(temp, t)
	}

	m := make(map[string]interface{})
	m["total"] = count
	m["data"] = temp
	bs, _ := json.Marshal(m)
	Reply(c, SlotReply{Data: bs})
}

//获取单个文章
func (a *ArticleService) ArticleGet(c *gin.Context) {
	id := c.Param("id") //拿到文章id，参数
	db := a.db.ConnectMysql().(*gorm.DB)
	rdb := a.rdb.ConnectRedis()

	//文章阅读量加1
	_, err := rdb.Do("incr", id)
	if err != nil {
		logger.Error(err)
		return
	}
	//获取当前文章的阅读量
	readCount, err := redis.Int(rdb.Do("get", id))
	if err != nil {
		logger.Error(err)
		return
	}

	//获取当前文章的评论量
	commentCount := 0
	var ac []models.ArticleComment
	db.Where("article_id = ?", id).Find(&ac).Count(&commentCount)

	//获取文章
	article := models.Article{}
	if err := db.Where("id = ?", id).Find(&article).Error; err != nil {
		logger.Error(err)
		return
	}
	article.ReadCount = uint(readCount)
	article.CommentCount = uint(commentCount)
	m := make(map[string]interface{})
	m["data"] = article
	bs, _ := json.Marshal(m)
	Reply(c, SlotReply{Data: bs})
}

//类型上传
func (a *ArticleService) ArticleTypeUpload(c *gin.Context) {
	data, _ := c.Get("data")
	tempData := data.(map[string]interface{})
	name := tempData["type_name"].(string)
	if name == "" {
		Reply(c, SlotReply{Error: "类型不能为空"})
		return
	}
	//将头字母转换成大写
	newName := Capitalize(name)

	//获取mysql连接对象
	db := a.db.ConnectMysql().(*gorm.DB)

	types := models.Type{}
	if err := db.Where("type_name = ?", newName).Find(&types).Error; err == nil {
		Reply(c, SlotReply{Error: "该类型已经存在"})
		return
	}
	types.CreatedAt = GetTime()
	types.TypeName = newName
	db.Save(&types)
	m := make(map[string]interface{})
	m["data"] = "添加成功"
	bs, _ := json.Marshal(m)
	Reply(c, SlotReply{Data: bs})
}

//类型全部获取
func (a *ArticleService) ArticleTypeGetAll(c *gin.Context) {
	db := a.db.ConnectMysql().(*gorm.DB)
	var types []models.Type
	if err := db.Find(&types).Error; err != nil {
		logger.Error(err)
		return
	}
	m := make(map[string]interface{})
	m["data"] = types
	bs, _ := json.Marshal(m)
	Reply(c, SlotReply{Data: bs})
}

//类型删除
func (a *ArticleService) ArticleTypeDel(c *gin.Context) {
	data, _ := c.Get("data")
	tempData := data.(map[string]interface{})
	typeId := tempData["id"].(float64)
	db := a.db.ConnectMysql().(*gorm.DB)

	//删除
	types := models.Type{}
	types.Id = uint(typeId)
	if err := db.Delete(&types).Error; err != nil {
		logger.Error(err)
		return
	}
	m := make(map[string]interface{})
	m["data"] = "删除成功"
	bs, _ := json.Marshal(m)
	Reply(c, SlotReply{Data: bs})
}
