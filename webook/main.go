package main

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"go_demo/webook/internal/repository"
	"go_demo/webook/internal/repository/dao"
	"go_demo/webook/internal/service"
	"go_demo/webook/internal/web"
	//"gorm.io/gorm"
	"github.com/gin-contrib/sessions"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"strings"
	"time"
)

func main() {
	db := initDB()
	server := initWebServer()
	u := initUser(db)
	u.RegisterRoutes(server)
	server.Run(":8080")
}

func initWebServer() *gin.Engine {
	server := gin.Default()
	//解决跨域问题
	server.Use(cors.New(cors.Config{
		AllowOrigins: []string{"http://localhost:3000"},
		AllowMethods: []string{"POST", "GET"},
		AllowHeaders: []string{"Content-Type", "Authorization"},
		//ExposeHeaders:    []string{"Content-Type", "Authorization"}, 允许带jtw-token
		//是否允许带cookie 即用户认证信息
		AllowCredentials: true,
		//允许的来源
		AllowOriginFunc: func(origin string) bool {
			if strings.HasPrefix(origin, "http://localhost") {
				return true
			}
			return strings.Contains(origin, "yourcampany.com")
		},
		MaxAge: 12 * time.Hour,
	})) //"root:root@tcp(localhost:13316)/webook"

	//sessions
	store := cookie.NewStore([]byte("secret"))
	server.Use(sessions.Sessions("mysession", store))

	return server
}

func initUser(db *gorm.DB) *web.UserHandler {
	ud := dao.NewUserDao(db)
	repo := repository.NewUserRepository(ud)
	svc := service.NewUserService(repo)
	u := web.NewUserHandler(svc)
	//c.RegisterRoutes(server)
	return u
}

func initDB() *gorm.DB {
	//server := gin.Default()
	//dsn := "root:root@tcp(localhost:13316)/webook"
	db, err := gorm.Open("mysql", "root:root@tcp(localhost:13316)/webook")
	if err != nil {
		//只在初始化过程中panic
		panic(err) //panic 相当于整个goroutine结束
	}
	err = dao.InitTable(db)
	if err != nil {
		panic(err)
	}
	return db
}
