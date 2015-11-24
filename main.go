// Go hello-world implementation for eleme/hackathon.

package main

import (
	"api"
	"fmt"
	"net/http"
	"os"
	"runtime"
)

func main() {
	runtime.GOMAXPROCS(8)
	api.Init()

	http.HandleFunc("/foods",api.FoodsHandler)
	http.HandleFunc("/login", api.LoginHandler)
	http.HandleFunc("/carts", api.CartsHandler)
	http.HandleFunc("/carts/", api.CartsAddFoodHandler)
	http.HandleFunc("/orders",api.OrderHandler)
	http.HandleFunc("/admin/orders",api.AdminGetOrderHandler)
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
