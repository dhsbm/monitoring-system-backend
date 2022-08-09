package main

import (
	"log"
	"server/cors"
	"server/logs"
	"server/report"
	"server/web"

	"net/http"
	"server/mydb"

	"server/user"

	"github.com/gin-gonic/gin"
)

func main() {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	r.Use(cors.Cors()) // 处理跨域
	defer mydb.DB.Close()
	r.GET("/", func(context *gin.Context) {
		log.Println("收到get请求")
		context.String(http.StatusOK, "服务器已正常开启")
	})

	r.POST("/user/login", user.Login)
	r.POST("/user/register", user.Register)
	r.GET("/user/info", user.GetInfo)

	r.GET("/web/list", web.GetWeb)
	r.POST("/web/add", web.AddWeb)
	r.PUT("/web/edit", web.EditWeb)
	r.DELETE("/web/delete", web.DeleteWeb)

	r.POST("/logs/all", logs.AllLogs)
	r.POST("/logs/stat", logs.StatLogs)
	r.POST("/logs/err", logs.ErrLogs)
	r.POST("/logs/per", logs.PerLogs)
	r.POST("/logs/beh", logs.BehLogs)
	r.POST("/logs/http", logs.HttpLogs)

	r.POST("/report/err", report.ReportErr)
	r.POST("/report/per", report.ReportPer)
	r.POST("/report/beh", report.ReportBeh) // 应该为beh
	r.POST("/report/http", report.ReportHttp)

	// go stat()
	r.Run(":3333")
}

// 每15分钟统计一次
// 报警功能——待完成
// func stat() {
// 	now := time.Now().Unix()
// 	gap := 1 * 60
// 	// gap := 15 * 60
// 	rest := int(now % int64(gap))
// 	time.Sleep(time.Second * time.Duration(gap-rest))
// 	for {
// 		log.Println(time.Now())
// 		// time.Sleep(time.Second * 1000)
// 		time.Sleep(time.Second * time.Duration(gap))
// 	}
// }
