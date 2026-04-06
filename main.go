package main

import (
	"bluebell/controllers"
	"bluebell/dao/minio"
	"bluebell/dao/pgsql"
	"bluebell/dao/redis"
	"bluebell/logger"
	"bluebell/pkg/snowflake"
	"bluebell/routes"
	"bluebell/setting"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"
)

// @title GamePulse
// @version 1.0
// @description 游戏脉络（游脉）社区，专注游戏社区舆情分析
// @termsOfService http://swagger.io/terms/

// @contact.name 张政
// @contact.url http://www.swagger.io/support
// @contact.email 1296775258@qq.com

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /api/v1
func main() {
	// 1. 加载配置
	if err := setting.Init(); err != nil {
		fmt.Printf("init setting failed error:%v\n", err)
		return
	}

	// 2. 初始化日志
	if err := logger.Init(setting.Conf.LogConfig, setting.Conf.Mode); err != nil {
		fmt.Printf("init setting failed error:%v\n", err)
		return
	}
	// 延迟注册一下，把缓冲区的文件追加到日志文件中
	defer zap.L().Sync()
	// zap.ReplaceGlobals(lg)后 通过zap.L()调用
	zap.L().Debug("logger init success...")

	// 3. 初始化MySQL连接
	if err := pgsql.Init(setting.Conf.PostgresConfig); err != nil {
		fmt.Printf("init setting failed error:%v\n", err)
		return
	}
	defer pgsql.Close()

	// 4. 初始化Redis连接
	if err := redis.Init(); err != nil {
		fmt.Printf("init setting failed error:%v\n", err)
		return
	}
	defer redis.Close()

	// 【新增】：初始化 MinIO
	if err := minio.Init(setting.Conf.MinIOConfig); err != nil {
		fmt.Printf("init minio failed, err:%v\n", err)
		zap.L().Fatal("init minio failed", zap.Error(err))
		return
	}
	// minio是无状态HTTP客户端 不需要defer关闭连接

	// 雪花算法生成id
	if err := snowflake.Init(setting.Conf.StartTime, setting.Conf.MachineID); err != nil {
		fmt.Printf("init snowflake failed error:%v\n", err)
		return
	}

	// 初始化gin框架内置的校验器使用的翻译器
	if err := controllers.InitTrans("zh"); err != nil {
		fmt.Printf("Init validator trans failed, err:%v\n", err)
		return
	}
	// 5. 注册路由
	r := routes.Setup(setting.Conf.Mode)
	// 6. 启动服务（优雅关机）
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", setting.Conf.Port),
		Handler: r,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	// 等待中断信号量来优雅关闭服务器，为关闭服务器设置一个5秒的超时
	quit := make(chan os.Signal, 1) // 创建一个接收信号的通道
	// kill 默认会发送syscall.SIGTERM信号
	// kill -2 发送 syscall.SIGINT 信号，我们常用的Ctrl+C就是触发系统的SIGINT信号
	// kill -9 发送 syscall.SIGKILL 信号，但是不能被捕获，所以不需要添加它
	// signal.Notify把收到的 syscall.SIGINT或syscall.SIGTERM 信号转发给quit
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM) // 此处不会阻塞
	<-quit                                               // 阻塞在此处，当接收到上述两种信号时才会往下执行
	zap.L().Info("Shutdown Server ...")
	// 创建一个5秒超时的context
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	// 5秒内优雅关闭服务（将未处理玩的请求处理完再关闭服务），超过5秒就超时退出
	if err := srv.Shutdown(ctx); err != nil {
		zap.L().Fatal("Server Shutdown", zap.Error(err))
	}

}
