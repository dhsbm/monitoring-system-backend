package web

// 网站接口

import (
	"log"
	"net/http"
	"server/mydb"
	"server/util"

	"github.com/gin-gonic/gin"
)

type Web struct {
	Web_id    int
	Title     string
	Url       string
	Warn_list string
}

// 获取网站列表接口
func GetWeb(c *gin.Context) {
	user_id, msg := util.CheckHeader(c)
	if msg != "" {
		c.JSON(http.StatusOK, gin.H{"code": 401, "msg": msg})
		return
	}
	rows, _ := mydb.DB.Query(
		"select web_id, title, url, warn_list from webs "+
			"where user_id=?",
		user_id)
	res := []Web{}
	for rows.Next() {
		var item Web
		rows.Scan(&item.Web_id, &item.Title, &item.Url, &item.Warn_list)
		res = append(res, item)
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "获取网站列表成功", "data": gin.H{"web_list": res}})
}

// 添加网站接口
func AddWeb(c *gin.Context) {
	user_id, msg := util.CheckHeader(c)
	if msg != "" {
		c.JSON(http.StatusOK, gin.H{"code": 401, "msg": msg})
		return
	}

	var json Web
	if err := c.ShouldBindJSON(&json); err != nil {
		// gin.H封装了生成json数据的工具
		c.JSON(http.StatusOK, gin.H{"code": 400, "msg": "参数不正确"})
		return
	}

	ret, err := mydb.DB.Exec(
		"insert into `webs` "+
			"(user_id, title, url) "+
			"values(?,?,?)",
		user_id, json.Title, json.Url)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "msg": "数据库故障"})
		log.Println("插入网站数据失败")
		return
	}

	id, _ := ret.LastInsertId()
	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "插入成功", "data": gin.H{"web_id": id}})
}

// 编辑网站接口
func EditWeb(c *gin.Context) {
	user_id, msg := util.CheckHeader(c)
	if msg != "" {
		c.JSON(http.StatusOK, gin.H{"code": 401, "msg": msg})
		return
	}
	var json Web
	if err := c.ShouldBindJSON(&json); err != nil {
		// gin.H封装了生成json数据的工具
		c.JSON(http.StatusOK, gin.H{"code": 400, "msg": "参数不正确"})
		return
	}

	_, err := mydb.DB.Exec("update webs set title=?, url=?, warn_list=? "+
		"where `user_id`=? AND `web_id`=?",
		json.Title, json.Url, json.Warn_list, user_id, json.Web_id)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 400, "msg": "修改失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "修改成功"})
}

// 删除网站接口
func DeleteWeb(c *gin.Context) {
	user_id, msg := util.CheckHeader(c)
	if msg != "" {
		c.JSON(http.StatusOK, gin.H{"code": 401, "msg": msg})
		return
	}
	var json Web
	if err := c.ShouldBindJSON(&json); err != nil {
		// gin.H封装了生成json数据的工具
		c.JSON(http.StatusOK, gin.H{"code": 400, "msg": "参数不正确"})
		return
	}

	_, err := mydb.DB.Exec("delete from webs where user_id=? and web_id=?",
		user_id, json.Web_id)
	if err == nil {
		c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "删除成功"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 400, "msg": "删除失败"})
}
