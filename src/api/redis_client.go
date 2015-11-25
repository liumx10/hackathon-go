package api

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"gopkg.in/redis.v3"
	"os"
)

var client_chan chan *redis.Client
var db *sql.DB

func BorrowClient() *redis.Client {
	return <-client_chan
}
func ReturnClient(client *redis.Client) {
	client_chan <- client
}

func InitNewRedisClient() {
	client_chan = make(chan *redis.Client, 128)
	for i := 0; i < 128; i++ {
		client := redis.NewClient(&redis.Options{
			Addr:     os.Getenv("REDIS_HOST") + ":" + os.Getenv("REDIS_PORT"),
			Password: "",
			DB:       0,
		})
		_, err := client.Ping().Result()
		if err != nil {
			fmt.Println("redis connect error")
		}
		//fmt.Println(pong, err)
		client_chan <- client
	}

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
