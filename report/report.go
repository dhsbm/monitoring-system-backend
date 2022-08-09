package report

// 上报接口

import (
	"log"
	"net/http"
	"server/mydb"
	"strconv"

	"github.com/gin-gonic/gin"
)

type Err struct {
	Web_id  int
	Url     string
	Time    int64
	Type    int
	Message string
	Stack   string
}

func ReportErr(c *gin.Context) {
	var json Err

	if err := c.ShouldBindJSON(&json); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 400, "msg": "参数不正确"})
		return
	}
	_, err := mydb.DB.Exec(
		"insert into `err_cache_logs` "+
			"(web_id, url, time, `type`, message, stack) "+
			"values(?,?,?,?,?,?)",
		json.Web_id, json.Url, json.Time, json.Type, json.Message, json.Stack)

	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "msg": "数据库故障"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "上报异常日志成功"})
}

type Per struct {
	Web_id int
	Url    string
	Time   int64
	Dns    int
	Fp     int
	Fcp    int
	Lcp    int
	Dcl    int
	L      int
}

// 上报异常日志
func ReportPer(c *gin.Context) {
	var json Per

	if err := c.ShouldBindJSON(&json); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 400, "msg": "参数不正确"})
		return
	}
	_, err := mydb.DB.Exec(
		"insert into `per_cache_logs` "+
			"(web_id, url, time, dns, fp, fcp, lcp, dcl, l) "+
			"values(?,?,?,?,?,?,?,?,?)",
		json.Web_id, json.Url, json.Time, json.Dns, json.Fp, json.Fcp, json.Lcp, json.Dcl, json.L)

	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "msg": "数据库故障"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "上报性能日志成功"})
}

type Beh struct {
	Web_id   int
	Url      string
	Time     int64
	Duration int
	Ip       string
	Area     int
	Brower   int
	User     int
}

type Obj struct {
	Browser int
	Area    int
}

// 上报行为日志
func ReportBeh(c *gin.Context) {
	var json Beh

	if err := c.ShouldBindJSON(&json); err != nil {
		log.Println(json)
		c.JSON(http.StatusOK, gin.H{"code": 400, "msg": "参数不正确"})
		return
	}
	_, err := mydb.DB.Exec(
		"insert into `beh_cache_logs` "+
			"(web_id, url, time, duration, ip, area, browser, user) "+
			"values(?,?,?,?,?,?,?,?)",
		json.Web_id, json.Url, json.Time, json.Duration, json.Ip, json.Area, json.Brower, json.User)

	str1, str2 := "browser"+strconv.Itoa(json.Brower), "area"+strconv.Itoa(json.Area)
	// log.Println(str1, str2)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "msg": "数据库故障"})
		return
	}

	// 往数据库记录用户浏览器和地区
	var obj Obj
	_ = mydb.DB.QueryRow(
		"select "+str1+","+str2+" from webs "+
			"where web_id=?",
		json.Web_id).Scan(&obj.Browser, &obj.Area)
	// log.Println(obj)

	mydb.DB.Exec(
		"update webs set "+str1+"=?, "+str2+"=? "+
			"where `web_id`=?",
		obj.Browser+1, obj.Area+1, json.Web_id)

	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "上报行为日志成功"})
}

type Http struct {
	Web_id   int
	Url      string
	Time     int64
	Send_url string
	Res_time int
	Way      string
	Success  int
	Status   int
	Res_body string
}

// 上报网络日志
func ReportHttp(c *gin.Context) {
	var json Http

	if err := c.ShouldBindJSON(&json); err != nil {
		// gin.H封装了生成json数据的工具
		log.Println(json)
		c.JSON(http.StatusOK, gin.H{"code": 400, "msg": "参数不正确"})
		return
	}
	_, err := mydb.DB.Exec(
		"insert into `http_cache_logs` "+
			"(web_id, url, time, send_url, res_time, way, success, `status`, res_body) "+
			"values(?,?,?,?,?,?,?,?,?)",
		json.Web_id, json.Url, json.Time, json.Send_url, json.Res_time, json.Way, json.Success, json.Status, json.Res_body)

	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "msg": "数据库故障"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "上报行为日志成功"})
}
