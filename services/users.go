package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	logger "github.com/zztroot/zztlog"
	"strings"
	"time"
	. "zzt_blog/common"
	"zzt_blog/databases"
	"zzt_blog/models"
	"github.com/levigross/grequests"
)

type UserService struct {
	db  databases.MysqlConn
	rdb databases.RedisConn
	mdb databases.MongoConn
}

//等三方登录github
func (u *UserService) GithubCallback(c *gin.Context){
	logger.Info("===============进入回调")
	//获取code
	code := c.Query("code")
	logger.Info("code：",code)
	url := "https://github.com/login/oauth/access_token"
	response, err := grequests.Post(url, &grequests.RequestOptions{
		JSON: map[string]string{
			"client_id":"4de0f5e10262e36b2223",
			"client_secret":"9cbfb994783eb2ee54e01c9da76208ce15d18240",
			"code":code,
			//"redirect_uri":"https://www.zztdd.cn/api/user/githubCallbackCode",
			"state":"STATE",
		},
	})
	logger.Info("access_token响应数据：", string(response.Bytes()))

	if err != nil {
		logger.Error(err)
		return
	}
	temp := strings.Split(string(response.Bytes()), "=")
	accessToken := strings.TrimRight(temp[1], "&scope")
	logger.Info("access_token号:", accessToken)

	//通过token获取用户数据
	getUrl := fmt.Sprintf(`https://api.github.com/user?access_token=%s`, accessToken)
	getResponse, err := grequests.Get(getUrl, &grequests.RequestOptions{
		Headers: map[string]string{
			"Authorization":fmt.Sprintf(`token %s`, accessToken),
		},
	})
	if err != nil {
		logger.Error(err)
		return
	}
	logger.Info("获取的用户数据:",string(getResponse.Bytes()))

	m := make(map[string]interface{})
	_ = json.Unmarshal(getResponse.Bytes(), &m)
	if m["login"] == nil {
		c.Redirect(200, "https://www.zztdd.cn")
		return
	}
	res := make(map[string]interface{})
	res["data"] = m
	c.Redirect(200, "https://www.zztdd.cn")
	bs, _ := json.Marshal(res)
	Reply(c, SlotReply{Data: bs})
}


type JWTClaims struct { // token里面添加用户信息，验证token后可能会用到用户信息
	jwt.StandardClaims
	UserID      uint     `json:"user_id"`
	Password    string   `json:"password"`
	Username    string   `json:"username"`
	FullName    string   `json:"full_name"`
	Permissions []string `json:"permissions"`
}

var (
	Secret     = "zzt" // 加密
	ExpireTime = 1800  // token有效期
)

func (u *UserService) QueryInfo(c *gin.Context) {
	db := u.db.ConnectMysql().(*gorm.DB)
	info := models.Info{}
	if err := db.First(&info).Error; err != nil {
		logger.Error(err)
		return
	}
	m := make(map[string]interface{})
	m["data"] = info
	bs, _ := json.Marshal(m)
	Reply(c, SlotReply{Data: bs})
}

func (u *UserService) UploadInfo(c *gin.Context) {
	db := u.db.ConnectMysql().(*gorm.DB)
	temp, _ := c.Get("data")
	data := temp.(map[string]interface{})

	if data["content"].(string) == "" {
		logger.Error("内容为空")
		return
	}
	info := models.Info{}
	if data["id"] == nil {
		info.Content = data["content"].(string)
		info.CreatedAt = GetTime()
		if err := db.Save(&info).Error; err != nil {
			logger.Error(err)
			return
		}
	}else {
		if err := db.Model(&info).Where("id = ?", data["id"].(float64)).Update("content", data["content"].(string)).Error; err != nil {
			logger.Error(err)
			return
		}
	}
	Reply(c, SlotReply{Error: ""})
}

func (u *UserService) Login(c *gin.Context) {
	tempData, _ := c.Get("data")
	data := tempData.(map[string]interface{})
	db := u.db.ConnectMysql().(*gorm.DB)
	userName := data["user_name"].(string)
	password := data["password"].(string)

	//判断数据是否为空
	if userName == "" || password == "" {
		logger.Error("数据不能为空")
		Reply(c, SlotReply{Error: "数据不能为空"})
		return
	}

	//判断账号密码的长度
	if len(userName) < 6 || len(password) < 6 {
		logger.Error("账号或密码长度不够")
		Reply(c, SlotReply{Error: "账号或密码长度不够"})
		return
	}

	//查询用户
	user := models.User{}
	if err := db.Where("user_name = ?", userName).Find(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			logger.Error("账号或密码错误")
			Reply(c, SlotReply{Error: "账号或密码错误，请重新输入"})
			return
		}else {
			logger.Error(err)
			return
		}
	}

	//判断账号是否正确
	if user.UserName != userName {
		logger.Error("账号或密码错误")
		Reply(c, SlotReply{Error: "账号或密码错误，请重新输入"})
		return
	}

	//判断密码是否正确
	if user.Pwd != password {
		logger.Error("账号或密码错误")
		Reply(c, SlotReply{Error: "账号或密码错误，请重新输入"})
		return
	}

	//登录成功，生产token
	claims := &JWTClaims{
		UserID:   user.Id,
		Username: userName,
		Password: password,
		//Permissions: []string{},
	}
	claims.IssuedAt = time.Now().Unix()
	claims.ExpiresAt = time.Now().Add(time.Second * time.Duration(ExpireTime)).Unix()
	signedToken, err := getToken(claims)
	if err != nil {
		logger.Error(err)
		return
	}

	re := make(map[string]interface{})
	re["data"] = map[string]string{
		"token": signedToken,
		"status":"successful",
	}
	bs, _ := json.Marshal(re)
	//返回结果
	Reply(c, SlotReply{Data: bs})
}

//获得token
func getToken(claims *JWTClaims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(Secret))
	if err != nil  {
		return "", errors.New("生成token错误")
	}
	return signedToken, nil
}
