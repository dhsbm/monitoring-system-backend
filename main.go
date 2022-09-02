package main

import (
	"log"
	"server/cors"
	"server/logs"
	"server/report"
	"server/web"
	"strconv"
	"time"

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

	go maintainDatabase()
	// go stat()
	r.Run(":3333")
}

var monthMs int64 = 30 * 24 * 60 * 60 * 1000

// 维护数据库数据，每天执行一次，将30天前的数据后移30天
func maintainDatabase() {
	now := time.Now().Unix()
	gap := 24 * 60 * 60 // 一天的秒数
	rest := int(now % int64(gap))
	time.Sleep(time.Second * time.Duration(gap-rest))

	for {
		log.Println("数据库开始更新")
		pre := time.Now().Unix()*1000 - monthMs
		update("beh_cache_logs", pre)
		update("per_cache_logs", pre)
		update("err_cache_logs", pre)
		update("http_cache_logs", pre)
		time.Sleep(time.Second * time.Duration(gap))
	}
}

func update(table string, time int64) {
	str := "UPDATE " + table +
		" SET time=time+" + strconv.FormatInt(monthMs, 10) +
		" where time < " + strconv.FormatInt(time, 10)

	_, err := mydb.DB.Exec(str)

	if err != nil {
		log.Printf(table + "更新出错")
	}
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
