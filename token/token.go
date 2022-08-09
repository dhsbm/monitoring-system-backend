package token

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt"
)

var jwtKey = []byte("my_secret_key")

type Claims struct {
	User_id int
	jwt.StandardClaims
}

// 根据id生成token
func GetToken(userId int) string {
	claims := &Claims{
		User_id: userId,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Unix() + int64(time.Hour*24*30),
			IssuedAt:  time.Now().Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString(jwtKey)
	return tokenString
}

// 解析token
func ParseToken(tokenString string) (int, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})

	if err != nil {
		// log.Println(err)
		return -1, errors.New("token已过期")
	}

	claims, ok := token.Claims.(*Claims)

	if ok && token.Valid {
		return claims.User_id, nil
	}

	return -1, errors.New("token解析出错")
}
