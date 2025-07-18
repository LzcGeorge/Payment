package main

import (
	"net/http"
	"strings"
	"time"
	"wepay/internal/repository"
	"wepay/internal/repository/dao"
	"wepay/internal/service"
	"wepay/internal/web"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	db := initDB()
	server := initWebServer()
	// appid, mchid, certificateSerialNo, privateKeyPath, wechatPayPublicKeyId, wechatPayPublicKeyPath
	client := web.NewClient("wxdsfahgf234sd", "123523432", "ajkhyuiKJSAHDn124fsadasda", "certs/private_key.pem", "adsbvcretgnfsde", "certs/public_key.pem")
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

func initTransfer(db *gorm.DB, client *web.Client) *web.TransferHandler {
	dao := dao.NewTransferDao(db)
	repo := repository.NewTransferRepository(dao)
	svc := service.NewTransferService(repo)
	return web.NewTransferHandler(svc, *client)
}
