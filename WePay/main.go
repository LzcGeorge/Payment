package main

import (
	"gorm.io/gorm/logger"
	"net/http"
	"strings"
	"time"
	"wepay/internal/repository"
	"wepay/internal/repository/dao"
	"wepay/internal/service"
	"wepay/internal/service/wxpay_utility"
	"wepay/internal/web"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	db := initDB()
	server := initWebServer()
	client := initClient()
	transferHandler := initTransfer(db, client)

	transferHandler.RegisterRoutes(server.Group("/transfer"))
	// 定义路由
	server.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Welcome to WePay API",
			"time":    time.Now().Format(time.RFC3339),
		})
	})

	_ = server.Run(":8080") // listen and serve on 8080
}

func initDB() *gorm.DB {
	db, err := gorm.Open(mysql.Open("root:root@tcp(localhost:13326)/wepay?charset=utf8mb4&parseTime=True&loc=Local"))
	if err != nil {
		panic(err)
	}
	err = dao.InitTable(db)
	if err != nil {
		panic(err)
	}
	db.Logger = logger.Default.LogMode(logger.Info)
	return db
}

func initWebServer() *gin.Engine {
	server := gin.Default()

	// middleware: 跨域请求
	server.Use(cors.New(cors.Config{
		AllowHeaders: []string{"Content-Type", "Authorization"},
		// 允许跨域请求携带 cookie
		AllowCredentials: true,
		AllowOriginFunc: func(origin string) bool {
			if strings.HasPrefix(origin, "http://localhost") {
				// 本地开发环境
				return true
			}
			return strings.Contains(origin, "yourcompany.com")
		},
		MaxAge: 12 * time.Hour,
	}))

	return server
}

func initClient() web.Client {
	mchConfig, err := wxpay_utility.CreateMchConfig(
		"1368139500",                // mchid
		"ajkhyuiKJSAHDn124fsadasda", // certificateSerialNo
		"certs/private_key.pem",     // privateKeyPath
		"adsbvcretgnfsde",           // wechatPayPublicKeyId
		"certs/public_key.pem",      // wechatPayPublicKeyPath
	)
	if err != nil {
		panic(err)
	}
	return web.NewClient(
		"wxb9f4f763e5d4a6de", // appid
		mchConfig,
		"http://wepay.selfknow.cn", // notifyUrl
	)
}

func initTransfer(db *gorm.DB, client web.Client) *web.TransferHandler {
	transferDao := dao.NewTransferDao(db)
	transferRepo := repository.NewTransferRepository(transferDao)
	transferSvc := service.NewTransferService(transferRepo)

	userDao := dao.NewUserDao(db)
	userRepo := repository.NewUserRepository(userDao)
	userSvc := service.NewUserService(userRepo)

	return web.NewTransferHandler(transferSvc, userSvc, client)
}
