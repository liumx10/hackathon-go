package api

import (
	"fmt"
	"os"
	"strconv"
)

type Food struct {
	Id    int `json:"id"`
	Price int `json:"price"`
	Stock int `json:"stock"`
}
type Foods struct {
	Foods []Food
}

func (foods *Foods) update(id int, price int, st int) {
	foods.Foods[id] = Food{id, price, st}
}
func (foods *Foods) add(id int, price int, stock int) {
	foods.Foods[id] = Food{id, price, stock}
}
func (foods *Foods) init() {
	foods.Foods = make([]Food, 128)
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
	for _, food := range foods.Foods {
		fmt.Println(food.Id, food.Price, food.Stock)
	}

	res := foods.check(1)
	fmt.Println("food id: 1 ", res)
	res = foods.check(100)
	fmt.Println("food id: 100 ", res)
	res = foods.check(101)
	fmt.Println("food id: 101 ", res)

	fmt.Println("check redis")
	price, _ := client.Get("food:1:price").Result()
	stock, _ := client.Get("food:1:stock").Result()
	fmt.Println("1: price:", price, " stock: ", stock)
	price, _ = client.Get("food:2:price").Result()
	stock, _ = client.Get("food:3:stock").Result()
	fmt.Println("2: price:", price, " stock: ", stock)
}
