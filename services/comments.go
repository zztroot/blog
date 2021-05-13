package services

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	logger "github.com/zztroot/zztlog"
	. "zzt_blog/common"
	"zzt_blog/databases"
	"zzt_blog/models"
)

type CommentService struct {
	db  databases.MysqlConn
	rdb databases.RedisConn
	mdb databases.MongoConn
}

//添加评论
func (this *CommentService) UploadComment(c *gin.Context){
	temp, _ := c.Get("data")
	data := temp.(map[string]interface{})
	db := this.db.ConnectMysql().(*gorm.DB)

	//判断评论内容是否为空
	if data["user_name"].(string) == "" || data["content"].(string) == "" {
		Reply(c, SlotReply{Error: "数据不能为空"})
		return
	}

	comment := models.WebComment{}
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

//查询评论
func (this *CommentService) QueryComment(c *gin.Context){
	db := this.db.ConnectMysql().(*gorm.DB)

	//查询
	var comments []models.WebComment
	if err := db.Order("created_at desc").Find(&comments).Error; err != nil {
		logger.Error(err)
		return
	}

	//返回结果
	m := make(map[string]interface{})
	m["data"] = comments
	bs, _ := json.Marshal(m)
	Reply(c, SlotReply{Data: bs})
}
