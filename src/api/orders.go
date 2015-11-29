package api

import (
	"io/ioutil"

	"gopkg.in/redis.v3"
	//	"log"
	"net/http"
	"strconv"
	"strings"
)

type OrderPostReply struct {
	Id string `json:"id"`
}

type OrderGetReplyItem struct {
	Id    string        `json:"id"`
	Items []interface{} `json:"items"`
	Total int           `json:"total"`
}

type MakeOrderArgs struct {
	CartId string `json:"cart_id"`
}

func OrderHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		Response(w, 400, Reply{"MALFORMED_JSON", "格式错误"})
		return
	}
	user, err := users.get_user_by_request(r)
	if err != nil {
		Response(w, 401, Reply{"INVALID_ACCESS_TOKEN", "无效的令牌"})
		return
	}
	user_id := strconv.Itoa(user.id)
	client := BorrowClient()
	defer ReturnClient(client)
	if r.Method == "POST" {
		r.ParseForm()

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			Error(w, err, "can not read body")
			return
		}
		if len(body) == 0 {
			Response(w, 400, Reply{"EMPTY_REQUEST", "请求体为空"})
			return
		}

		var t MakeOrderArgs
		err = Parser(body, &t)

		if err != nil {
			Response(w, 400, Reply{"MALFORMED_JSON", "格式错误"})
			return
		}

		var ismember, cart_exist *redis.BoolCmd
		var user_cart_id *redis.StringCmd
		var cart_foodcmd *redis.StringSliceCmd

		init_pip := client.Pipeline()

		ismember = init_pip.SIsMember("ALL_CARTS", t.CartId)
		cart_exist = init_pip.Exists(user_id + ":order")
		user_cart_id = init_pip.Get(user_id + ":carts")
		cart_foodcmd = init_pip.LRange(t.CartId+":cart_foods", 0, 2)
		init_pip.Exec()
		defer init_pip.Close()

		//log.Println(ismember.Val(), ", ", cart_exist.Val(), ", ", user_cart_id.Val(), ", ", cart_foodcmd.Val())
		//if !client.SIsMember("ALL_CARTS", t.CartId).Val() {
		if !ismember.Val() {
			Response(w, 404, Reply{"CART_NOT_FOUND", "篮子不存在"})
			return
		}
		//user_cart_id := client.Get(user_id + ":carts").Val()
		//if t.CartId != user_cart_id {
		if t.CartId != user_cart_id.Val() {
			Response(w, 401, Reply{"NOT_AUTHORIZED_TO_ACCESS_CART", "无权限访问指定的篮子"})
			return
		}

		//if client.Exists(user_id + ":order").Val() {
		if cart_exist.Val() {
			Response(w, 403, Reply{"ORDER_OUT_OF_LIMIT", "每个用户只能下一单"})
			return
		}
		cart_foods := cart_foodcmd.Val()
		//cart_foods := client.LRange(t.CartId+":cart_foods", 0, 2).Val()


		food_ids:=make([]string,len(cart_foods))
		food_counts:=make([]int,len(cart_foods))
		food_stock:=make([]int,len(cart_foods))

		for i := 0; i < len(cart_foods); i++ {
			strs := strings.FieldsFunc(cart_foods[i], func(s rune) bool {
				return s == ':'
			})
			food_ids[i] = strs[0]
			food_counts[i], _ = strconv.Atoi(strs[1])
		}

		pipeline := client.Pipeline()

		decr_cmd:=make([]*redis.IntCmd,len(cart_foods))
		for i := 0; i < len(cart_foods); i++ {
			decr_cmd[i] = pipeline.DecrBy("food:"+food_ids[i]+":stock", int64(food_counts[i]))
		}
		pipeline.Exec()
		pipeline.Close()
		rollback := false

		for i := 0; i < len(cart_foods); i++ {
			res := decr_cmd[i].Val()
			if res < 0 {
				rollback = true
				break
			}
		}

		if rollback {
			rollback_pipeline := client.Pipeline()
			for i := 0; i < len(cart_foods); i++ {
				rollback_pipeline.IncrBy("food:"+food_ids[i]+":stock", int64(food_counts[i]))
			}
			rollback_pipeline.Exec()
			rollback_pipeline.Close()
			Response(w, 403, Reply{"FOOD_OUT_OF_STOCK", "食物库存不足"})
			return
		}

		order_content := ""
		for i := 0; i < len(cart_foods); i++ {
			order_content += food_ids[i] + ":" + strconv.Itoa(food_counts[i])
			if i != len(order_content)-1 {
				order_content += ","
			}
		}

		finish_pipeline := client.Pipeline()
		finish_pipeline.Set(user_id+":order", order_content, 0)
		finish_pipeline.Set(user_id+":order_id", t.CartId, 0)
		finish_pipeline.SAdd("ALL_ORDERS", order_content+";"+user_id+";"+t.CartId)
		finish_pipeline.Exec()
		finish_pipeline.Close()
		for i := 0; i < len(cart_foods); i++ {
			id, _ := strconv.Atoi(food_ids[i])
			foods.update(id, food_stock[i]-food_counts[i])
		}

		Response(w, 200, OrderPostReply{t.CartId})

	} else if r.Method == "GET" {
		r.ParseForm()

		order_content, err := client.Get(user_id + ":order").Result()

		if err != nil {
			Response(w, 200, EmptyReply{})
			return
		}

		order_id := client.Get(user_id + ":order_id").Val()

		cart_foods := strings.FieldsFunc(order_content, func(s rune) bool {
			return s == ','
		})

		total := 0
		var replyItem OrderGetReplyItem
		replyItem.Id = order_id
		replyItem.Items = make([]interface{}, len(cart_foods))
		for i := 0; i < len(cart_foods); i++ {
			strs := strings.FieldsFunc(cart_foods[i], func(s rune) bool {
				return s == ':'
			})
			food_id, _ := strconv.Atoi(strs[0])
			count, _ := strconv.Atoi(strs[1])
			data := make(map[string]int)
			data["food_id"] = food_id
			data["count"] = count
			replyItem.Items[i] = data
			total += count * foods.get_price(food_id)
		}
		replyItem.Total = total
		var reply [1]OrderGetReplyItem
		reply[0] = replyItem
		Response(w, 200, reply)

	}
}
