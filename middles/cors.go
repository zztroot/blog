package middles

import (
	"errors"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"github.com/levigross/grequests"
	logger "github.com/zztroot/zztlog"
	"net/http"
	"time"
	"zzt_blog/common"
	"zzt_blog/models"
)

func Cors() gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method

		c.Header("Access-Control-Allow-Origin", "*") // 可将将 * 替换为指定的域名
		c.Header("Access-Control-Allow-Methods", "POST, GET, PUT, DELETE, UPDATE")
		c.Header("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept, Authorization")
		c.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Cache-Control, Content-Language, Content-Type")
		c.Header("Access-Control-Allow-Credentials", "true")

		if method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
		}
		c.Next()
	}
}

var (
	Secret     = "zzt" // 加密
)
type JWTClaims struct { // token里面添加用户信息，验证token后可能会用到用户信息
	jwt.StandardClaims
	UserID      uint     `json:"user_id"`
	Password    string   `json:"password"`
	Username    string   `json:"username"`
	//Permissions []string `json:"permissions"`
}

//token验证
func TokenCheck() gin.HandlerFunc{
	return func(c *gin.Context) {
		strToken := c.Request.Header.Get("authorization")
		//strToken := c.Param("token")
		claim,err := verifyAction(strToken)
		if err != nil {
			//c.JSON(http.StatusOK, gin.H{
			//	"error":"无效token",
			//})
			common.Reply(c, common.SlotReply{Error: "无效token"})
			logger.Error(err)
			c.Abort()
			return
		}
		c.Set("token_info", claim)
		c.Next()
	}
}

func verifyAction(strToken string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(strToken, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(Secret), nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*JWTClaims)
	if !ok {
		return nil, errors.New("jwt错误")
	}
	if err := token.Claims.Valid(); err != nil {
		return nil, err
	}
	return claims, nil
}

//向百度蜘蛛提交url
func PushBaiduUrl(dbs interface{}){
	for {
		logger.Info("进入百度蜘蛛提交url")
		db := dbs.(*gorm.DB)
		var articles []models.Article
		if err := db.Where("delete_at = ?", "no").Find(&articles).Error; err != nil {
			logger.Error(err)
			return
		}

		var urls []string
		//for _, v := range articles {
		//	urls = append(urls, fmt.Sprintf(`https://www.zztdd.cn/showArticle?id=%d`, v.Id))
		//}
		urls = append(urls, "https://www.zztdd.cn/")
		urls = append(urls, "https://www.zztdd.cn/tools")
		//urls = append(urls, "https://zztdd.cn/tools_html/pdfToWord.html")
		//urls = append(urls, "https://zztdd.cn/tools_html/pdfToImg.html")
		//urls = append(urls, "https://zztdd.cn/tools_html/wordToPdf.html")
		//urls = append(urls, "https://zztdd.cn/tools_html/htmlToPdf.html")

		//将文章推送给百度收录
		for _, v := range urls {
			url := "http://data.zz.baidu.com/urls?site=https://www.zztdd.cn&token=uvqkZC9WzyCGL1lJ"
			re, err := grequests.Post(url, &grequests.RequestOptions{
				Headers: map[string]string{"Content-Type": "text/plain"},
				JSON: v,
			})
			if err != nil {
				logger.Error(err)
				return
			}
			logger.Info(string(re.Bytes()))
		}

		//等待时间
		time.Sleep(time.Hour*24)
	}
}

type CommentHost struct {
	HostList map[string]string
}

func (this *CommentHost)CommentVerify() gin.HandlerFunc {
	return func(c *gin.Context) {
		value, ok := this.HostList[c.Request.Host]
		if !ok {
			this.HostList[c.Request.Host] = common.GetTimeGzh()
		}else {
			nowTime := time.Now().Add(-time.Second).Format("2006-01-02 15:04:05")
			if value >= nowTime {
				common.Reply(c, common.SlotReply{Error: "requests timeout"})
				c.Abort()
				return
			}
		}
		c.Next()
	}
}