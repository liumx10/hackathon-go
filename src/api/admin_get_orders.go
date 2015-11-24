package api

import (
	"net/http"
	"strconv"
	"strings"
)

type AdminOrderGetReplyItem struct {
	Id    string        `json:"id"`
	UserId int		`json:"user_id"`
	Items []interface{} `json:"items"`
	Total int           `json:"total"`
}
func AdminGetOrderHandler(w http.ResponseWriter, r *http.Request){
		
		err := r.ParseForm()
		if err != nil {
			Response(w, 400, Reply{"MALFORMED_JSON", "格式错误"})
			return
		}
		_,err = users.get_user_by_request(r)
		if err != nil {
			Response(w, 401, Reply{"INVALID_ACCESS_TOKEN", "无效的令牌"})
			return
		}

		client:=BorrowClient()
		defer ReturnClient(client)
		orders :=client.SMembers("ALL_ORDERS").Val()
		
		reply:=make([]AdminOrderGetReplyItem,len(orders))
		
		for i:=0;i<len(orders);i++{
			
			strs := strings.FieldsFunc(orders[i], func(s rune) bool {
				return s == ';'
			})
			order_content:=strs[0]
			user_id,_:=strconv.Atoi(strs[1])
			order_id := strs[2]
			cart_foods := strings.FieldsFunc(order_content, func(s rune) bool {
				return s == ','
			})
	
			total := 0
			var replyItem AdminOrderGetReplyItem
			replyItem.UserId = user_id
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
			reply[i]=replyItem
		}

		Response(w, 200, reply)
}
