package controller

import (
	. "app/entity"
	"fmt"
	"net/http"
	"time"
	. "web-api-ex/common"
	v2 "web-api-ex/controller/v2"

	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	jsoniter "github.com/json-iterator/go"
	"go.uber.org/zap"
)

var signKey = "6553795CC3FA62E6E62875C48DB35463"

// 路由注册，及绑定访问权限
func Register(g *gin.Engine, method, url string, cb gin.HandlerFunc, accTypes []ACCOUNT_TYPE) {
	UrlPloys().SetPloy(url, accTypes)
	if accTypes != nil {
		switch method {
		case "GET", "get":
			g.GET(url, CheckPermissions(), cb)
		case "POST", "post":
			g.POST(url, CheckPermissions(), cb)
		case "PUT", "put":
			g.PUT(url, CheckPermissions(), cb)
		}
	} else {
		switch method {
		case "GET", "get":
			g.GET(url, cb)
		case "POST", "post":
			g.POST(url, cb)
		case "PUT", "put":
			g.PUT(url, cb)
		}
	}
}

// 身份验证
func CheckPermissions() gin.HandlerFunc {
	return func(c *gin.Context) {
		// zap.L().Debug("======>request", zap.Any("uri", c.Request.URL.Path))
		token := c.Query("token")
		if token == "" {
			token = c.PostForm("token")
		}
		if token == "" {
			c.String(http.StatusForbidden, "没有权限！")
			c.Abort()
			return
		}
		//
		if c.Request.URL.Path == "/api/auth/v2/game/gameUrl" {
			return
		}
		if c.Request.URL.Path == "/api/auth/v2/queryOrder" {
			roundId := c.Query("roundId")
			account := c.Query("account")
			if token != HmacSha256ToHex(signKey, []byte(fmt.Sprintf("%s%s", roundId, account))) {
				c.String(http.StatusForbidden, "没有权限！")
				c.Abort()
				return
			}
		} else {
			str := AesDecrypt(AesKey, AesIV, token)
			loginInfo := &LoginInfo{}
			jsoniter.UnmarshalFromString(str, loginInfo)
			if loginInfo.LoginTime+TokenTimeOut < time.Now().Unix() {
				c.Redirect(http.StatusPermanentRedirect, "/login")
				return
			}
			if loginInfo.Id <= 0 || !UrlPloys().CheckPloy(c.Request.URL.Path, int(loginInfo.UserType)) {
				c.String(http.StatusForbidden, "没有权限！")
				c.Abort()
				return
			}
			c.Set("agentId", int64(loginInfo.AgentId))
			c.Set("id", int64(loginInfo.Id))
		}
	}
}

// 初始化gin
func NewRouter() *gin.Engine {
	engine := gin.New()
	engine.Use(ginzap.RecoveryWithZap(zap.L(), true))
	g := engine
	//设置跨域
	g.Use(Cors())

	Register(g, "get", "/api/auth/v2/queryOrder", v2.QueryOrder, []ACCOUNT_TYPE{M, A, S})
	Register(g, "get", "/api/auth/v2/game/gameUrl", v2.GameUrls, []ACCOUNT_TYPE{M, A})

	return engine
}
