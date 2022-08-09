package logs

// 日志接口
import (
	"bytes"
	"net/http"
	"server/mydb"
	"server/util"
	"strings"

	"github.com/gin-gonic/gin"
)

type AllJson struct {
	Web_id   int
	End_time int64
}

// 请求概览页的所有信息
func AllLogs(c *gin.Context) {
	// 验证token
	if _, msg := util.CheckHeader(c); msg != "" {
		c.JSON(http.StatusOK, gin.H{"code": 401, "msg": msg})
		return
	}
	var json AllJson
	if err := c.ShouldBindJSON(&json); err != nil {
		// gin.H封装了生成json数据的工具
		c.JSON(http.StatusOK, gin.H{"code": 400, "msg": "参数不正确"})
		return
	}
	gap := int64(24 * 3600 * 1000)
	startTime := json.End_time - 7*24*3600*1000
	err := getStat(json.Web_id, 0, startTime, json.End_time, gap)
	per := getStat(json.Web_id, 1, startTime, json.End_time, gap)
	user1 := getStat(json.Web_id, 2, startTime, json.End_time, gap)
	user2 := getStat(json.Web_id, 3, startTime, json.End_time, gap)
	http1 := getStat(json.Web_id, 4, startTime, json.End_time, gap)
	http2 := getStat(json.Web_id, 5, startTime, json.End_time, gap)
	browser, area := getBrowserAndArea(json.Web_id)
	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "获取图表信息成功", "data": gin.H{
		"err": err, "per": per, "user1": user1, "user2": user2,
		"http1": http1, "http2": http2, "browser": browser, "area": area}})
}

type StatJson struct {
	Web_id   int   `json:"web_id"`   // 网站id
	Kind     int   `json:"kind"`     // 类别，异常/性能/行为/请求：0-3
	Time     int   `json:"time"`     // 时间维度，0-4对应4h/1d/7d/14d/30d
	End_time int64 `json:"end_time"` // 结束时间戳
	Index    int   `json:"index"`    // 索引，仅在日志类型为行为或请求时生效，因为它们有两个表，所以要区分
}

// 请求统计(图表)信息
func StatLogs(c *gin.Context) {
	// 验证token
	if _, msg := util.CheckHeader(c); msg != "" {
		c.JSON(http.StatusOK, gin.H{"code": 401, "msg": msg})
		return
	}

	var json StatJson
	if err := c.ShouldBindJSON(&json); err != nil {
		// gin.H封装了生成json数据的工具
		c.JSON(http.StatusOK, gin.H{"code": 400, "msg": "参数不正确"})
		return
	}

	startTime := json.End_time
	index := -1
	switch json.Kind {
	case 0:
		index = 0
	case 1:
		index = 1
	case 2:
		if json.Index == 0 {
			index = 2
		} else {
			index = 3
		}
	case 3:
		if json.Index == 0 {
			index = 4
		} else {
			index = 5
		}
	}
	// 4h/1d/7d/14d/30d
	// 15m/2h//1d/1d/2d
	gap := int64(0)
	switch json.Time {
	case 0:
		gap = 15 * 60 * 1000
		startTime -= 4 * 3600 * 1000
	case 1:
		gap = 2 * 3600 * 1000
		startTime -= 24 * 3600 * 1000
	case 2:
		gap = 24 * 3600 * 1000
		startTime -= 7 * 24 * 3600 * 1000
	case 3:
		gap = 24 * 3600 * 1000
		startTime -= 14 * 24 * 3600 * 1000
	case 4:
		gap = 2 * 24 * 3600 * 1000
		startTime -= 30 * 24 * 3600 * 1000
	}

	res := getStat(json.Web_id, index, startTime, json.End_time, gap)

	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "获取统计信息成功", "data": res})
}

type ErrLog struct {
	Log_id  int    `json:"log_id"`
	Web_id  int    `json:"web_id"`
	Url     string `json:"url"`
	Time    int64  `json:"time"`
	Type    int    `json:"type"`
	Message string `json:"message"`
	Stack   string `json:"stack"`
}
type ErrCondition struct {
	Url     string
	Time    string
	Type    string
	Message string
}
type ErrJson struct {
	Web_id    int
	Page      int
	Condition ErrCondition
}

// 请求异常日志列表
func ErrLogs(c *gin.Context) {
	// 验证token
	if _, msg := util.CheckHeader(c); msg != "" {
		c.JSON(http.StatusOK, gin.H{"code": 401, "msg": msg})
		return
	}

	var json ErrJson
	if err := c.ShouldBindJSON(&json); err != nil {
		// gin.H封装了生成json数据的工具
		c.JSON(http.StatusOK, gin.H{"code": 400, "msg": "参数不正确"})
		return
	}

	rows, _ := mydb.DB.Query(
		"select log_id, url, time, type, message, stack from err_cache_logs "+
			"where web_id=?"+
			likeCondition(json.Condition.Url, "Url")+
			rangeCondition(json.Condition.Time, "Time")+
			screenCondition(json.Condition.Type, "Type")+
			likeCondition(json.Condition.Message, "Message")+
			" order by time DESC",
		json.Web_id)

	var total, now int
	res := []ErrLog{}
	for rows.Next() {
		if now >= json.Page*10-10 && now < json.Page*10 {
			var item ErrLog
			rows.Scan(&item.Log_id, &item.Url, &item.Time, &item.Type, &item.Message, &item.Stack)
			res = append(res, item)
		}
		total++
		now++
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "获取异常日志成功",
		"data": gin.H{"total": total, "page": json.Page, "size": 10, "logs": res}})
}

type PerLog struct {
	Log_id int    `json:"log_id"`
	Web_id int    `json:"web_id"`
	Url    string `json:"url"`
	Time   int64  `json:"time"`
	Dns    int    `json:"dns"`
	Fp     int    `json:"fp"`
	Fcp    int    `json:"fcp"`
	Lcp    int    `json:"lcp"`
	Dcl    int    `json:"dcl"`
	L      int    `json:"l"`
}
type PerCondition struct {
	Url  string
	Time string
	Dns  string
	Fp   string
	Fcp  string
	Lcp  string
	Dcl  string
	L    string
}
type PerJson struct {
	Web_id    int
	Page      int
	Condition PerCondition
}

// 请求性能日志列表
func PerLogs(c *gin.Context) {
	// 验证token
	if _, msg := util.CheckHeader(c); msg != "" {
		c.JSON(http.StatusOK, gin.H{"code": 401, "msg": msg})
		return
	}

	var json PerJson
	if err := c.ShouldBindJSON(&json); err != nil {
		// gin.H封装了生成json数据的工具
		c.JSON(http.StatusOK, gin.H{"code": 400, "msg": "参数不正确"})
		return
	}

	rows, _ := mydb.DB.Query("select log_id, url, time, dns, fp, fcp, lcp, dcl, l from per_cache_logs "+
		"where web_id=?"+
		likeCondition(json.Condition.Url, "Url")+
		rangeCondition(json.Condition.Time, "Time")+
		rangeCondition(json.Condition.Dns, "Dns")+
		rangeCondition(json.Condition.Fp, "Fp")+
		rangeCondition(json.Condition.Fcp, "Fcp")+
		rangeCondition(json.Condition.Lcp, "Lcp")+
		rangeCondition(json.Condition.Dcl, "Dcl")+
		rangeCondition(json.Condition.L, "L")+
		" order by time DESC",
		json.Web_id)

	var total, now int
	res := []PerLog{}
	for rows.Next() {
		if now >= json.Page*10-10 && now < json.Page*10 {
			var item PerLog
			rows.Scan(&item.Log_id, &item.Url, &item.Time, &item.Dns, &item.Fp, &item.Fcp, &item.Lcp, &item.Dcl, &item.L)
			res = append(res, item)
		}
		total++
		now++
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "获取性能日志成功",
		"data": gin.H{"total": total, "page": json.Page, "size": 10, "logs": res}})
}

type BehLog struct {
	Log_id   int    `json:"log_id"`
	Web_id   int    `json:"web_id"`
	Url      string `json:"url"`
	Time     int64  `json:"time"`
	Duration int    `json:"duration"`
	Ip       string `json:"ip"`
	Area     int    `json:"Area"`
	Browser  int    `json:"Browser"`
	User     int    `json:"user"`
}
type BehCondition struct {
	Url      string
	Time     string
	Duration string
	Ip       string
}
type BehJson struct {
	Web_id    int
	Page      int
	Condition BehCondition
}

// 请求行为日志列表
func BehLogs(c *gin.Context) {
	// 验证token
	if _, msg := util.CheckHeader(c); msg != "" {
		c.JSON(http.StatusOK, gin.H{"code": 401, "msg": msg})
		return
	}

	var json BehJson
	if err := c.ShouldBindJSON(&json); err != nil {
		// gin.H封装了生成json数据的工具
		c.JSON(http.StatusOK, gin.H{"code": 400, "msg": "参数不正确"})
		return
	}

	rows, _ := mydb.DB.Query("select log_id, url, time, duration, ip, area from beh_cache_logs "+
		"where web_id=?"+
		likeCondition(json.Condition.Url, "Url")+
		rangeCondition(json.Condition.Time, "Time")+
		rangeCondition(json.Condition.Duration, "Duration")+
		likeCondition(json.Condition.Ip, "Ip")+
		" order by time DESC",
		json.Web_id)

	var total, now int
	res := []BehLog{}
	for rows.Next() {
		if now >= json.Page*10-10 && now < json.Page*10 {
			var item BehLog
			rows.Scan(&item.Log_id, &item.Url, &item.Time, &item.Duration, &item.Ip, &item.Area)
			res = append(res, item)
		}
		total++
		now++
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "获取行为日志成功",
		"data": gin.H{"total": total, "page": json.Page, "size": 10, "logs": res}})
}

type HttpLog struct {
	Log_id   int    `json:"log_id"`
	Web_id   int    `json:"web_id"`
	Url      string `json:"url"`
	Time     int64  `json:"time"`
	Send_url string `json:"send_url"`
	Res_time int    `json:"res_time"`
	Way      string `json:"way"`
	Success  int    `json:"success"`
	Status   int    `json:"status"`
	Res_body string `json:"res_body"`
}
type HttpCondition struct {
	Url      string
	Time     string
	Res_time string
	Way      string
	Send_url string
	Success  string
}
type HttpJson struct {
	Web_id    int
	Page      int
	Condition HttpCondition
}

// 请求网络日志列表
func HttpLogs(c *gin.Context) {
	// 验证token
	if _, msg := util.CheckHeader(c); msg != "" {
		c.JSON(http.StatusOK, gin.H{"code": 401, "msg": msg})
		return
	}

	var json HttpJson
	if err := c.ShouldBindJSON(&json); err != nil {
		// gin.H封装了生成json数据的工具
		c.JSON(http.StatusOK, gin.H{"code": 400, "msg": "参数不正确"})
		return
	}

	rows, _ := mydb.DB.Query("select log_id, url, time, send_url, res_time, way, success, `status`, res_body from http_cache_logs "+
		"where web_id=?"+
		likeCondition(json.Condition.Url, "Url")+
		rangeCondition(json.Condition.Time, "Time")+
		rangeCondition(json.Condition.Res_time, "Res_time")+
		likeCondition(json.Condition.Way, "Way")+
		likeCondition(json.Condition.Send_url, "Send_url")+
		screenCondition(json.Condition.Success, "Success")+
		" order by time DESC",
		json.Web_id)

	var total, now int
	res := []HttpLog{}
	for rows.Next() {
		if now >= json.Page*10-10 && now < json.Page*10 {
			var item HttpLog
			rows.Scan(&item.Log_id, &item.Url, &item.Time, &item.Send_url,
				&item.Res_time, &item.Way, &item.Success, &item.Status, &item.Res_body)
			res = append(res, item)
		}
		total++
		now++
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "获取网络日志成功",
		"data": gin.H{"total": total, "page": json.Page, "size": 10, "logs": res}})
}

// 处理范围条件
func rangeCondition(str, key string) string {
	var buf bytes.Buffer
	if str != "" {
		strs := strings.Split(str, "_")
		buf.WriteString(" and ")
		buf.WriteString(key)
		buf.WriteString(">=")
		buf.WriteString(strs[0])
		buf.WriteString(" and ")
		buf.WriteString(key)
		buf.WriteString("<=")
		buf.WriteString(strs[1])
	}
	return buf.String()
}

// 处理匹配条件
func likeCondition(str, key string) string {
	var buf bytes.Buffer
	if str != "" {
		buf.WriteString(" and ")
		buf.WriteString(key)
		buf.WriteString(" like '%")
		buf.WriteString(str)
		buf.WriteString("%'")
	}
	return buf.String()
}

// 处理筛选条件
func screenCondition(str, key string) string {
	var buf bytes.Buffer
	if str != "" {
		buf.WriteString(" and ")
		buf.WriteString(key)
		buf.WriteString("=")
		buf.WriteString(str)
	}
	return buf.String()
}

// 获取统计信息函数，供接口调用
func getStat(id, index int, statrTime, endTime, gap int64) [][]int {
	res := [][]int{}
	len := int((endTime - statrTime) / gap)
	for i := 0; i < len; i++ {
		res = append(res, []int{})
	}

	switch index {
	case 0: // 异常
		{
			for i := 0; i < len; i++ {
				res[i] = []int{0, 0, 0, 0}
			}
			rows, _ := mydb.DB.Query(
				"select time, type from err_cache_logs "+
					"where web_id=?"+
					" and time>=? and time<?",
				id, statrTime, endTime)
			i := 0
			for rows.Next() {
				var item ErrLog
				rows.Scan(&item.Time, &item.Type)
				for !(item.Time >= statrTime && item.Time < statrTime+gap) {
					i++
					statrTime += gap
				}
				res[i][item.Type]++
			}
		}
	case 1: // 性能
		{
			for i := 0; i < len; i++ {
				res[i] = []int{0, 0, 0, 0, 0, 0}
			}
			rows, _ := mydb.DB.Query(
				"select time, dns, fp, fcp, lcp, dcl, l from per_cache_logs "+
					"where web_id=?"+
					" and time>=? and time<?",
				id, statrTime, endTime)

			i, count := 0, 0
			for rows.Next() {
				var item PerLog
				rows.Scan(&item.Time, &item.Dns, &item.Fp, &item.Fcp, &item.Lcp, &item.Dcl, &item.L)
				for !(item.Time >= statrTime && item.Time < statrTime+gap) {
					if count != 0 {
						util.ExceptAll(res[i], count)
					}
					i++
					count = 0
					statrTime += gap
				}
				res[i][0] += item.Dns
				res[i][1] += item.Fp
				res[i][2] += item.Fcp
				res[i][3] += item.Lcp
				res[i][4] += item.Dcl
				res[i][5] += item.L
				count++
			}
			if count != 0 {
				util.ExceptAll(res[i], count)
			}
		}
	case 2: //行为 pv/uv
		{
			for i := 0; i < len; i++ {
				res[i] = []int{0, 0}
			}
			rows, _ := mydb.DB.Query(
				"select time, user from beh_cache_logs "+
					"where web_id=?"+
					" and time>=? and time<?",
				id, statrTime, endTime)

			i := 0
			for rows.Next() {
				var item BehLog
				rows.Scan(&item.Time, &item.User)
				for !(item.Time >= statrTime && item.Time < statrTime+gap) {
					i++
					statrTime += gap
				}
				if item.User != 2 {
					res[i][1]++
				}
				res[i][0]++
			}
		}
	case 3: //行为 dration
		{
			for i := 0; i < len; i++ {
				res[i] = []int{0}
			}
			rows, _ := mydb.DB.Query(
				"select time, duration from beh_cache_logs "+
					"where web_id=?"+
					" and time>=? and time<?",
				id, statrTime, endTime)
			i, count := 0, 0
			for rows.Next() {
				var item BehLog
				rows.Scan(&item.Time, &item.Duration)
				for !(item.Time >= statrTime && item.Time < statrTime+gap) {
					if count != 0 {
						util.ExceptAll(res[i], count)
					}
					i++
					count = 0
					statrTime += gap
				}
				res[i][0] += item.Duration
				count++
			}
			if count != 0 {
				util.ExceptAll(res[i], count)
			}
		}
	case 4: //网络 成功失败数
		{
			for i := 0; i < len; i++ {
				res[i] = []int{0, 0}
			}
			rows, _ := mydb.DB.Query(
				"select time, success from http_cache_logs "+
					"where web_id=?"+
					" and time>=? and time<?",
				id, statrTime, endTime)

			i := 0
			for rows.Next() {
				var item HttpLog
				rows.Scan(&item.Time, &item.Success)
				for !(item.Time >= statrTime && item.Time < statrTime+gap) {
					i++
					statrTime += gap
				}
				res[i][item.Success]++
			}
		}
	case 5: //网络 响应时间
		{
			for i := 0; i < len; i++ {
				res[i] = []int{0}
			}
			rows, _ := mydb.DB.Query(
				"select time, res_time from http_cache_logs "+
					"where web_id=?"+
					" and time>=? and time<?",
				id, statrTime, endTime)
			i, count := 0, 0
			for rows.Next() {
				var item HttpLog
				rows.Scan(&item.Time, &item.Res_time)
				for !(item.Time >= statrTime && item.Time < statrTime+gap) {
					if count != 0 {
						util.ExceptAll(res[i], count)
					}
					i++
					count = 0
					statrTime += gap
				}
				res[i][0] += item.Res_time
				count++
			}
			if count != 0 {
				util.ExceptAll(res[i], count)
			}
		}
	}
	// log.Println(res)
	return res
}

type Web struct {
	Web_id                                                                                 int
	Browser0, Browser1, Browser2, Browser3, Browser4, Browser5, Browser6, Area0, Area1     int
	Area2, Area3, Area4, Area5, Area6, Area7, Area8, Area9, Area10, Area11, Area12, Area13 int
	Area14, Area15, Area16, Area17, Area18, Area19, Area20, Area21, Area22, Area23         int
	Area24, Area25, Area26, Area27, Area28, Area29, Area30, Area31, Area32, Area33, Area34 int
}

// 请求浏览器和地区信息
func getBrowserAndArea(id int) ([7]int, [35]int) {
	var web Web
	_ = mydb.DB.QueryRow(
		"select Browser0, Browser1, Browser2, Browser3, Browser4, Browser5, Browser6, "+
			"Area0, Area1, Area2, Area3, Area4, Area5, Area6, Area7, Area8, Area9, "+
			"Area10, Area11, Area12, Area13, Area14, Area15, Area16, Area17, Area18, "+
			"Area19, Area20, Area21, Area22, Area23, Area24, Area25, Area26, Area27, "+
			"Area28, Area29, Area30, Area31, Area32, Area33, Area34 from webs "+
			"where web_id=?",
		id).Scan(&web.Browser0, &web.Browser1, &web.Browser2, &web.Browser3, &web.Browser4, &web.Browser5, &web.Browser6,
		&web.Area0, &web.Area1, &web.Area2, &web.Area3, &web.Area4, &web.Area5, &web.Area6, &web.Area7, &web.Area8, &web.Area9,
		&web.Area10, &web.Area11, &web.Area12, &web.Area13, &web.Area14, &web.Area15, &web.Area16, &web.Area17, &web.Area18,
		&web.Area19, &web.Area20, &web.Area21, &web.Area22, &web.Area23, &web.Area24, &web.Area25, &web.Area26, &web.Area27,
		&web.Area28, &web.Area29, &web.Area30, &web.Area31, &web.Area32, &web.Area33, &web.Area34)

	browser := [7]int{web.Browser0, web.Browser1, web.Browser2, web.Browser3, web.Browser4, web.Browser5, web.Browser6}
	area := [35]int{web.Area0, web.Area1, web.Area2, web.Area3, web.Area4, web.Area5, web.Area6, web.Area7, web.Area8,
		web.Area9, web.Area10, web.Area11, web.Area12, web.Area13, web.Area14, web.Area15, web.Area16, web.Area17,
		web.Area18, web.Area19, web.Area20, web.Area21, web.Area22, web.Area23, web.Area24, web.Area25, web.Area26,
		web.Area27, web.Area28, web.Area29, web.Area30, web.Area31, web.Area32, web.Area33, web.Area34}

	return browser, area
}
