package routes

import (
	"gamepulse/controllers"
	"gamepulse/logger"
	"gamepulse/middlewares"
	"net/http"
	"time"

	_ "gamepulse/docs" // 千万不要忘了导入把你上一步生成的docs

	// 注意：这里换成了 github.com/swaggo/files
	swaggerFiles "github.com/swaggo/files"
	gs "github.com/swaggo/gin-swagger"

	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
)

func Setup(mode string) *gin.Engine {
	if mode == gin.ReleaseMode {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.New()
	r.Use(logger.GinLogger(), logger.GinRecovery(true), middlewares.RateLimitMiddleware(time.Second, 100))

	r.LoadHTMLFiles("./templates/index.html")
	r.Static("/assets", "./assets")
	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", nil) // 单页面应用，通过JS去改变页面上的内容;展示数据从后端来
	})

	r.GET("/swagger/*any", gs.WrapHandler(swaggerFiles.Handler))
	r.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})
	v1 := r.Group("/api/v1")
	v1.POST("/signup", controllers.SignUpHandler)
	v1.POST("/login", controllers.LoginHandler)

	v1.Use(middlewares.JWTAuthMiddleware()) // 应用JWT认证中间件

	{
		v1.GET("/community", controllers.CommunityHandler)
		v1.GET("/community/:id", controllers.CommunityDetailHandler)

		v1.POST("/post", controllers.CreatePostHandler)
		v1.GET("/post/:id", controllers.GetPostDetailHandler)
		v1.GET("/posts", controllers.GetListDetailHandler)
		v1.GET("/posts2", controllers.GetListDetailHandler2)

		v1.POST("/vote", controllers.PostVoteController)

		v1.DELETE("/post/:id", controllers.DeletePostHandler)

		v1.POST("/upload", controllers.UploadImageHandler)
	}
	pprof.Register(r) // 注册pprof相关路由
	r.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"msg": "no Route!!! 404",
		})
	})
	return r
}
