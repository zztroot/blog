package gzhapi

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"github.com/levigross/grequests"
	logger "github.com/zztroot/zztlog"
	"io/ioutil"
	"time"
	. "zzt_blog/common"
	"zzt_blog/databases"
	"zzt_blog/models"
)

type Gzh struct {
	db            databases.MysqlConn
	gzhTokenValue string
}

func (g *Gzh) handlerGzhToken() {
	db := g.db.ConnectMysql().(*gorm.DB)
	token := models.Gzh{}
	if err := db.Find(&token).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			logger.Error(err)
			gzhToken, err := g.getGzhToken()
			if err != nil {
				logger.Error(err)
				return
			}
			token.CreatedAt = GetTimeGzh()
			token.AccessToken = gzhToken["access_token"].(string)
			token.ExpiresIn = uint(gzhToken["expires_in"].(float64))
			if err := db.Create(&token).Error; err != nil {
				logger.Error(err)
				return
			}
			g.gzhTokenValue = gzhToken["access_token"].(string)
			return
		} else {
			logger.Error(err)
			return
		}
	}
	hour := getHourDiffer(token.CreatedAt, "2021-05-06 15:10:46")
	if hour >= 2 {
		gzhToken, err := g.getGzhToken()
		if err != nil {
			logger.Error(err)
			return
		}
		token.CreatedAt = GetTimeGzh()
		token.AccessToken = gzhToken["access_token"].(string)
		token.ExpiresIn = uint(gzhToken["expires_in"].(float64))
		if err := db.Save(&token).Error; err != nil {
			logger.Error(err)
			return
		}
		g.gzhTokenValue = gzhToken["access_token"].(string)
		return
	}
	g.gzhTokenValue = token.AccessToken
}

//获取相差时间
func getHourDiffer(startTime, endTime string) int64 {
	var hour int64
	t1, err := time.ParseInLocation("2006-01-02 15:04:05", startTime, time.Local)
	t2, err := time.ParseInLocation("2006-01-02 15:04:05", endTime, time.Local)
	if err == nil && t1.Before(t2) {
		diff := t2.Unix() - t1.Unix()
		hour = diff / 3600
		return hour
	} else {
		return hour
	}
}

//请求token
func (g *Gzh) getGzhToken() (map[string]interface{}, error) {
	const (
		GrantType = "client_credential"
		AppId     = "wxae5ca11d812b74f7"
		Secret    = "a278fc289c72e0768a96d9ea9cc9cf4d"
	)
	url := fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/token?grant_type=%s&appid=%s&secret=%s", GrantType, AppId, Secret)
	get, err := grequests.Get(url, &grequests.RequestOptions{
		Host: "172.16.102.128",
	})
	if err != nil {
		logger.Error(err)
		return nil, err
	}
	m := make(map[string]interface{})
	_ = json.Unmarshal(get.Bytes(), &m)
	return m, nil
}

//校验微信接口
func (g *Gzh) CheckSignature(c *gin.Context) {
	logger.Info("进入微信公众号接口验证")
	signature := c.Query("signature")
	timestamp := c.Query("timestamp")
	nonce := c.Query("nonce")
	echostr := c.Query("echostr")
	newStr := "1033141032" + timestamp + nonce
	logger.Info(newStr)
	logger.Info("接受的signature:", signature)
	tempSignature := ToSha1String(newStr)
	logger.Info("我转换的signature:", tempSignature)
	//a := "1033141032" +"1620222580"+"355601369"
	//zztlog.Debug(ToSha1String(a))
	if tempSignature != signature {
		Reply(c, SlotReply{Error: "signature失败"})
		return
	}
	c.String(200, echostr)
	//c.JSON(200, gin.H{
	//	"echostr":echostr,
	//})
}

//测试token
func (g *Gzh) TestToken(c *gin.Context) {
	g.handlerGzhToken()
	logger.Info(g.gzhTokenValue)
}

//创建公众号菜单
func (g *Gzh) CreateMenu(c *gin.Context) {
	g.handlerGzhToken()
	data, err := ioutil.ReadFile("./gzhApi/gzhMenu.json")
	if err != nil {
		logger.Info(err)
		return
	}
	m := make(map[string]interface{})
	_ = json.Unmarshal(data, &m)
	url := fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/menu/create?access_token=%s", g.gzhTokenValue)
	post, err := grequests.Post(url, &grequests.RequestOptions{
		JSON: m,
	})
	if err != nil {
		logger.Info(err)
		return
	}
	logger.Info(string(post.Bytes()))
}
