package api

import (
	"fmt"
	"gopkg.in/redis.v3"
	"os"
)

var client *redis.Client
func InitNewRedisClient(){
	client = redis.NewClient(&redis.Options{
		Addr:os.Getenv("REDIS_HOST")+":"+os.Getenv("REDIS_PORT"),
		Password:"",
		DB:0,
	})
	pong ,err :=client.Ping().Result()
	fmt.Println(pong,err)
} 