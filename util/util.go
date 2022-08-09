package util

// 工具包

import (
	"server/token"

	"github.com/gin-gonic/gin"
)

// 检查请求头token
func CheckHeader(c *gin.Context) (int, string) {
	tokenString := c.GetHeader("Authorization")
	if tokenString == "" {
		return -1, "用户未登录"
	}
	user_id, err := token.ParseToken(tokenString)
	if err != nil {
		return -1, "用户token失效"
	}

	return user_id, ""
}

// 三元表达式
func Ternary[T interface{}](sign bool, value1, value2 T) T {
	if sign {
		return value1
	} else {
		return value2
	}
}

// 切片的每一项除以一个数
func ExceptAll(slice []int, num int) {
	for i, l := 0, len(slice); i < l; i++ {
		slice[i] /= num
	}
}
