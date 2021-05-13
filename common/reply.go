package common

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/wonderivan/logger"
)

type SlotReply struct {
	Error interface{}
	Data  []byte
}

//json数据返回
func Reply(c *gin.Context, reply SlotReply) {
	if reply.Error == nil {
		m := make(map[string]interface{})
		err := json.Unmarshal(reply.Data, &m)
		if err != nil {
			logger.Error(err)
		}
		if m["total"] != nil {
			c.JSON(200, gin.H{"data": m["data"], "error": "", "total":m["total"]})
			return
		}
		c.JSON(200, gin.H{"data": m["data"], "error": ""})
		return
	}
	c.JSON(200, gin.H{"data": "", "error": reply.Error})
	return
}
