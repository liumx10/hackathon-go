package main

import (
	"fmt"
	"gopkg.in/redis.v3"
)

func main() {
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	pong, err := client.Ping().Result()
	fmt.Println(pong, err)

	if err != nil {
		return
	}

	err1 := client.Set("key", "value", 0).Err()
	if err1 != nil {
		panic(err)
	}

	val, err2 := client.Get("key").Result()
	if err2 != nil {
		panic(err)
	}
	fmt.Println("key", val)

	val2, err3 := client.Get("key2").Result()
	if err3 == redis.Nil {
		fmt.Println("key2 does not exists")
	} else if err != nil {
		panic(err)
	} else {
		fmt.Println("key2", val2)
	}

}
