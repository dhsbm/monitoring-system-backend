package user

// 用户接口
import (
	"log"
	"net/http"
	"server/mydb"
	"server/token"
	"server/util"

	"github.com/gin-gonic/gin"
)

type User struct {
	User_id  int
	Email    string
	Password string
	Name     string
}

// 登录接口
func Login(c *gin.Context) {
	var json User

	if err := c.ShouldBindJSON(&json); err != nil {
		// gin.H封装了生成json数据的工具
		c.JSON(http.StatusOK, gin.H{"code": 400, "msg": "参数不正确"})
		return
	}

	var data User
	err := mydb.DB.QueryRow(
		"select user_id,password,name from users "+
			"where email=?",
		json.Email).
		Scan(&data.User_id, &data.Password, &data.Name)
	if err != nil {
		log.Println("登录时查询数据失败")
		c.JSON(http.StatusOK, gin.H{"code": 500, "msg": "数据库故障"})
		return
	}

	if data.Password == json.Password {
		tokenString := token.GetToken(data.User_id)
		c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "登录成功", "data": gin.H{"user_id": data.User_id, "name": data.Name, "token": tokenString}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 400, "msg": "密码或用户名错误"})
}

// 注册接口
func Register(c *gin.Context) {
	var json User

	if err := c.ShouldBindJSON(&json); err != nil {
		// gin.H封装了生成json数据的工具
		c.JSON(http.StatusOK, gin.H{"code": 400, "msg": "参数不正确"})
		return
	}

	rows, _ := mydb.DB.Query("select user_id,password from users "+
		"where email=?",
		json.Email)
	for rows.Next() {
		c.JSON(http.StatusOK, gin.H{"code": 400, "msg": "邮箱已注册"})
		return

	}

	ret, err := mydb.DB.Exec(
		"insert into `users` "+
			"(password, email, name) "+
			"values(?,?,?) limit 1",
		json.Password, json.Email, json.Name)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "msg": "数据库故障"})
		log.Println("插入用户数据失败")
		return
	}
	id, _ := ret.LastInsertId()
	tokenString := token.GetToken(int(id))
	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "注册成功", "data": gin.H{"user_id": id, "name": json.Name, "token": tokenString}})
}

// 根据token请求用户信息接口
func GetInfo(c *gin.Context) {
	user_id, msg := util.CheckHeader(c)
	if msg != "" {
		c.JSON(http.StatusOK, gin.H{"code": 401, "msg": msg})
		return
	}

	rows, _ := mydb.DB.Query("select name from users where user_id=?", user_id)
	for rows.Next() {
		var data User
		rows.Scan(&data.Name)
		tokenString := token.GetToken(user_id)
		c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "获取用户信息成功", "data": gin.H{"user_id": user_id, "name": data.Name, "token": tokenString}})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 401, "msg": "用户token失效"})
}
