package mydb

// 启动mysql数据库
import (
	"database/sql"
	"log"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

var DB *sql.DB

func init() {
	db, err := sql.Open("mysql", "root:123456@tcp(localhost:3306)/mydb")
	if err != nil {
		log.Println("连接数据库异常")
		panic(err)
	}
	// 最大空闲连接数 默认为2
	db.SetMaxIdleConns(5)
	// 最大连接数，默认不限制
	db.SetMaxOpenConns(100)
	// 连接最大存活时间
	db.SetConnMaxLifetime(time.Minute * 3)
	// 空闲连接最大存活时间
	db.SetConnMaxIdleTime(time.Minute * 1)
	err = db.Ping()
	if err != nil {
		log.Println("数据库无法连接")
		db.Close()
		panic(err)
	}
	DB = db
}
