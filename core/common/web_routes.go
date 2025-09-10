package common

import (
	"BaseGoUni/core/api"
	"BaseGoUni/core/utils"
	_ "BaseGoUni/docs" // 导入生成的docs
	"context"
	"encoding/json"
	"fmt"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"io/ioutil"
	"log"
	"mime"
	"net/http"
	"strings"
	"time"
)

func InitGin() {
	gin.DefaultWriter = ioutil.Discard
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowAllOrigins = true
	corsConfig.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"}
	corsConfig.AllowHeaders = []string{"Content-Type", "X-XSRF-TOKEN", "Accept", "X-Requested-With", "Origin", "Content-Length", "Content-Type", "Authorization"}
	corsConfig.AllowCredentials = true
	corsConfig.MaxAge = 12 * time.Hour
	router.Use(cors.New(corsConfig))
	//docs.SwaggerInfo.Title = "rcs服务api"
	//docs.SwaggerInfo.Description = "rcs服务api"
	//docs.SwaggerInfo.Version = "1.0"
	//docs.SwaggerInfo.Host = "localhost:8080"
	//docs.SwaggerInfo.BasePath = "{{host}}"
	//docs.SwaggerInfo.Schemes = []string{"http", "https"}
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	_ = mime.AddExtensionType(".js", "application/javascript")
	router.Use(static.ServeRoot("/", "dist"))
	router.Use(hostInfoMiddleware())
	apiGroup := router.Group("/api/v1")
	{
		apiGroup.GET("/heath/check", heathCheck)
		apiGroup.POST("/user/award", api.AwardUser) // 内部用户余额变动
		apiGroup.POST("/user/login", api.UserLogin) // 管理员登录
	}
	// 通用接口
	commonGroup := router.Group("/api/v1/outside")
	commonGroup.Use(authMiddleware([]int{1, 2, 3, 4}, false, true))
	{
		commonGroup.GET("/user/unbind/gauth", api.UnBindGAuth) // 解绑谷歌验证
		commonGroup.PUT("/user/pass", api.ChangePass)          // 修改密码
		commonGroup.POST("/menus", api.GetMenus)
		commonGroup.GET("/routes", api.GetRoutes)
		commonGroup.POST("/user", api.GetUsers)
		commonGroup.POST("/roles", api.GetRoles)
		commonGroup.POST("/userInfo", api.CurrentUserInfo)
		commonGroup.POST("/resetPwd", api.ResetPassword)
		commonGroup.GET("/roleIds/:userId", api.GetRoleIds)
		commonGroup.POST("/role-menu", api.GetRoleMenus)
		commonGroup.GET("/role-menu-ids/:roleId", api.GetRoleMenuIds)
	}
	// 管理员接口
	manageGroup := router.Group("/api/v1/manager")
	manageGroup.Use(authMiddleware([]int{1, 2}, false, true))
	{
		manageGroup.POST("/setRole", api.SetRole)
		manageGroup.DELETE("/role/:id", api.DelRole)
		manageGroup.POST("/setUser", api.SetUser)
		manageGroup.POST("/delUsers", api.DelUsers)
		commonGroup.POST("/cashHistory", api.UserCashHistory)
		manageGroup.PUT("/setMenus", api.SetMenus)
		manageGroup.DELETE("/menu/:id", api.DelMenu)
		manageGroup.POST("/user/award", api.AdminAwardUser) // 用户余额管理(送钱/扣钱)
	}
	// 超级管理员接口
	adminGroup := router.Group("/api/v1/admin")
	adminGroup.Use(authMiddleware([]int{1}, false, true))
	{
		adminGroup.PUT("/menus", api.SetMenus)
		adminGroup.PUT("/role", api.SetRole)
		adminGroup.DELETE("/role/:id", api.DelRole)
		adminGroup.PUT("/user", api.SetUser)
	}
	{
		adminGroup.POST("/host_infos", api.GetHostInfos)
		adminGroup.PUT("/host_info", api.SetHostInfo)
		adminGroup.DELETE("/host_info/:id", api.DelHostInfo)
	}
	log.Printf("Start server at %s:%d ", utils.GlobalConfig.Host, utils.GlobalConfig.Port)
	apiURL := fmt.Sprintf("%s:%d", utils.GlobalConfig.Host, utils.GlobalConfig.Port)
	err := router.Run(apiURL)
	if err != nil {
		log.Printf("Init gin error.err=%v\n", err)
		return
	}
}

func hostInfoMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Printf("%s %s", c.Request.Method, c.Request.RequestURI)
		host := utils.GetRequestHost(c)
		hostInfo := utils.GetTempHostInfo(host)
		if hostInfo.ID == 0 {
			hostInfoStr, _ := json.Marshal(hostInfo)
			log.Printf("host:%s;hostInfo:%s", host, string(hostInfoStr))
			c.String(http.StatusNotFound, "")
			c.Abort()
			return
		}
		db := utils.NewPrefixDb(hostInfo.TablePrefix)
		c.Set("hostInfo", hostInfo)
		c.Set("db", db)
		c.Next()
	}
}

// 通行用户类型 / 是否单点登录 / 是否过滤特殊token
func authMiddleware(types []int, singleLogin bool, passChild bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			utils.UnauthorizedBack(c, "Authorization header is missing")
			c.Abort()
			return
		}
		authHeader = strings.TrimPrefix(authHeader, "Bearer ")
		hostInfo := utils.GetTempHostInfo(utils.GetRequestHost(c))
		userId, userType, hostName, childCode, _ := utils.ParseToken(utils.CsConfig.DefaultHost.AccessSecret, authHeader)
		if passChild && childCode != "" {
			utils.UnauthorizedBack(c, "token is invalid -1")
			c.Abort()
			return
		}
		if hostInfo.HostName != hostName {
			utils.UnauthorizedBack(c, "token is invalid 0")
			c.Abort()
			return
		}
		if userId == 0 {
			utils.UnauthorizedBack(c, "token is invalid 1")
			c.Abort()
			return
		}
		inType := false
		for _, tempUserType := range types {
			if userType == tempUserType {
				inType = true
				break
			}
		}
		if !inType {
			utils.UnauthorizedBack(c, "not support api")
			c.Abort()
			return
		}
		user := utils.GetTempUser(hostInfo.TablePrefix, userId)
		if !user.Enabled {
			utils.UnauthorizedBack(c, "token is invalid 2")
			c.Abort()
			return
		}
		if singleLogin {
			key := utils.KeySingle + utils.MD5(fmt.Sprintf("%d", userId))
			data := utils.RD.Get(context.Background(), key)
			if data == nil || data.Err() != nil {
				utils.UnauthorizedBack(c, "token is passed")
				c.Abort()
				return
			}
			if data.Val() != authHeader {
				utils.UnauthorizedBack(c, "already logout")
				c.Abort()
				return
			}
		}
		requestKey := utils.MD5(fmt.Sprintf("%s_%s", c.Request.Method, c.Request.RequestURI))
		lockKey := fmt.Sprintf(utils.KeyLockRequest, userId, requestKey)
		lock, _ := utils.AcquireLock(lockKey, 1*time.Second)
		if !lock {
			c.JSON(http.StatusBadRequest, gin.H{"error": "request too fast.Please try again later."})
			return
		}
		c.Set("childCode", childCode)
		c.Set("userId", userId)
		c.Set("userType", userType)
		c.Set("token", authHeader)
		c.Next()
		endTime := time.Now()
		latencyTime := endTime.Sub(startTime)
		if latencyTime > 1*time.Second {
			log.Printf("Request %s %s took %v", c.Request.Method, c.Request.URL, latencyTime)
		}
	}
}

func heathCheck(c *gin.Context) {
	requestHost := utils.GetRequestHost(c)
	log.Printf("Request url: %s", requestHost)
	log.Printf("Request Host: %s", c.Request.Host)
	c.String(200, "ok")
}
