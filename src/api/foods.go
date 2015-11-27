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
	Foods   map[int]Food
	FoodsId []int
	count   int
	m       *sync.RWMutex
}

func (foods *Foods) update(id int, st int) {
	food := foods.Foods[id]
	foods.Foods[id] = Food{id, food.Price, st}
}
func (foods *Foods) add(id int, price int, stock int) {
	foods.FoodsId = append(foods.FoodsId, id)
	foods.Foods[id] = Food{id, price, stock}
	foods.count++
}
func (foods *Foods) init() {
	foods.m = new(sync.RWMutex)
	foods.FoodsId = make([]int, 0, 100)
	foods.Foods = make(map[int]Food)
	foods.count = 0
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
			for i := range foods.FoodsId {
				stock, _ := client.Get("food:" + strconv.Itoa(i) + ":stock").Result()
				st, _ := strconv.Atoi(stock)
				foods.update(i, st)
			}
			foods.m.Unlock()
			time.Sleep(1000 * time.Millisecond)
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
	list := make([]Food, 0, 100)
	for i := range foods.FoodsId {
		list = append(list, foods.Foods[foods.FoodsId[i]])
	}
	Response(w, 200, list[0:foods.count])
	foods.m.RUnlock()
}
