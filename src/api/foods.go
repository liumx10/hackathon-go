package api

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"
)

type Food struct {
	Id    int `json:"id"`
	Price int `json:"price"`
	Stock int `json:"stock"`
}
type Foods struct {
	Foods []Food
	m     *sync.RWMutex
}

func (foods *Foods) update(id int, st int) {
	foods.Foods[id].Stock = st
}
func (foods *Foods) add(id int, price int, stock int) {
	foods.Foods[id] = Food{id, price, stock} //, new(sync.RWMutex)}
}
func (foods *Foods) init() {
	foods.Foods = make([]Food, 128)
	foods.m = new(sync.RWMutex)
}
func (foods *Foods) check(id int) bool {
	return id < 101 && id > 0
}
func (foods *Foods) get_price(id int) int {
	if id > 101 || id < 1 {
		return -1
	} else {
		return foods.Foods[id].Price
	}
}

var foods *Foods

func InitFood() {
	foods = new(Foods)
	foods.init()
	rows, err := db.Query("select * from food")
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
	defer rows.Close()
	client := BorrowClient()
	defer ReturnClient(client)
	for rows.Next() {
		var id, price, stock int
		rows.Scan(&id, &stock, &price)
		foods.add(id, price, stock)
		err := client.Set("food:"+strconv.Itoa(id)+":stock", strconv.Itoa(stock), 0).Err()
		if err != nil {
			fmt.Println(err)
		}
		err = client.Set("food:"+strconv.Itoa(id)+":price", strconv.Itoa(price), 0).Err()
		if err != nil {
			fmt.Println(err)
		}
	}

	res := foods.check(1)
	fmt.Println("food id: 1 ", res)
	res = foods.check(100)
	fmt.Println("food id: 100 ", res)
	res = foods.check(101)
	fmt.Println("food id: 101 ", res)

	go func() {
		client := BorrowClient()
		defer ReturnClient(client)
		for {
			foods.m.Lock()
			for i := 1; i < 101; i++ {
				stock, _ := client.Get("food:" + strconv.Itoa(i) + ":stock").Result()
				//		foods.Foods[i].m.Lock()
				foods.Foods[i].Stock, _ = strconv.Atoi(stock)
				//		foods.Foods[i].m.Unlock()
			}
			foods.m.Unlock()
			time.Sleep(1000 * time.Millisecond)
			fmt.Println("tick")
		}
	}()
}

func FoodsHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		Response(w, 400, Reply{"MALFORMED_JSON", "格式错误"})
		return
	}
	_, err = users.get_user_by_request(r)
	if err != nil {
		fmt.Println(err)
		Response(w, 401, Reply{"INVALID_ACCESS_TOKEN", "无效的令牌"})
		return
	}
	//fmt.Println("foods: 42 ", foods.Foods[42].Stock)
	foods.m.RLock()
	Response(w, 200, foods.Foods[1:101])
	foods.m.RUnlock()
}
