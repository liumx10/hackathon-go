// Go hello-world implementation for eleme/hackathon.

package main

import (
	"net/http"
	"runtime"
	"api"
	"os"
	"fmt"
)


func main() {
	runtime.GOMAXPROCS(4)
	api.InitNewRedisClient()
	http.HandleFunc("/carts",api.CardsHandler)
	http.HandleFunc("/carts/",api.CardsAddFoodHandler)
	host := os.Getenv("APP_HOST")
	port := os.Getenv("APP_PORT")
	if host == "" {
		host = "localhost"
	}
	if port == "" {
		port = "8080"
	}

	addr := fmt.Sprintf("%s:%s", host, port)
	http.ListenAndServe(addr, nil)
}
