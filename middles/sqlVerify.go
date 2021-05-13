package middles

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	logger "github.com/zztroot/zztlog"
	"io/ioutil"
	"strings"
	."zzt_blog/common"
)

//数据过滤
func FilterSql() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == "GET" {
			c.Next()
			return
		}
		data := make(map[string]interface{})
		_ = c.BindJSON(&data)
		m := data
		sql, _ := ioutil.ReadFile("verify_sql.json")
		temp := make(map[string]interface{})
		_ = json.Unmarshal(sql, &temp)
		sqlList := temp["verify_sql"].([]interface{})
		for _, v := range m {
			switch v.(type) {
			case string:
				for _, v1 := range sqlList {
					if strings.Contains(v.(string), v1.(string)) {
						logger.Error(fmt.Sprintf(`非法参数`))
						err := ErrParameterInvalid
						Reply(c, SlotReply{Error: err})
						c.Abort()
						return
					}
				}
			default:
				continue
			}
		}
		c.Set("data", data)
		c.Next()
	}
}

