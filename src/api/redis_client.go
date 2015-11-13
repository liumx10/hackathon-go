package api

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"gopkg.in/redis.v3"
	"os"
)

var client *redis.Client
var db *sql.DB

func InitNewRedisClient() {
	client = redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_HOST") + ":" + os.Getenv("REDIS_PORT"),
		Password: "",
		DB:       0,
	})
	pong, err := client.Ping().Result()
	fmt.Println(pong, err)
}

func InitNewMysqlClient() {
	hostname := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	dbname := os.Getenv("DB_NAME")
	username := os.Getenv("DB_USER")
	pwd := os.Getenv("DB_PASS")

	str := username + ":" + pwd + "@tcp(" + hostname + ":" + port + ")/" + dbname
	fmt.Println(str)
	db, _ = sql.Open("mysql", str)

	error := db.Ping()
	fmt.Println(error)
}
