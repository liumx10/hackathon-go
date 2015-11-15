package api

import (
	"strconv"
	"strings"
	"net/http"
	"io/ioutil"
)

type OrderPostReply struct{
	Id		string `json:"id"`
}

type OrderGetReplyItem struct{
	Id		string `json:"id"`
	Items 	[]interface{} `json:"items"`
	Total	int `json:"total"`
}


type MakeOrderArgs struct{
	CartId    string `json:"cart_id"`
}

func OrderHandler(w http.ResponseWriter, r *http.Request){
	err := r.ParseForm();
	if err!=nil {
		Response(w, 400, Reply{"MALFORMED_JSON", "格式错误"})
		return
	}
	user,err:=users.get_user_by_request(r)
	if err!=nil{
		Response(w, 401, Reply{"INVALID_ACCESS_TOKEN", "无效的令牌"})
		return
	}
	user_id :=strconv.Itoa(user.id)
	if r.Method=="POST"{
		r.ParseForm();
		
		
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
	
		if err!=nil {
			Response(w, 400, Reply{"MALFORMED_JSON", "格式错误"})
			return
		}

		
		
		
		
		if !client.SIsMember("ALL_CARTS",t.CartId).Val(){
			Response(w,404,Reply{"CART_NOT_FOUND","篮子不存在"})
			return
		}
		user_cart_id := client.Get(user_id+":carts").Val()
		if t.CartId != user_cart_id{
			Response(w,403,Reply{"NOT_AUTHORIZED_TO_ACCESS_CART","无权限访问指定的篮子"})
			return
		}
		
		if client.Exists(user_id+":order").Val() {
			Response(w,403,Reply{"ORDER_OUT_OF_LIMIT","每个用户只能下一单"})
			return
		}
		
		cart_foods:=client.LRange(t.CartId+":cart_foods",0,2).Val()
		
		ok := false
		multi := client.Multi()
		defer multi.Close()
		var food_ids [3]string
		var food_counts [3] int
		var food_stock [3] int
		discarded:= false
		for !ok{
			for i:=0;i<len(cart_foods);i++{
				strs := strings.FieldsFunc(cart_foods[i],func(s rune) bool{
					return s==':'
				})
				food_ids[i] = strs[0]
				food_counts[i],_ = strconv.Atoi(strs[1])
				multi.Watch("food:"+food_ids[i]+":stock")
			}
			
			for i:=0;i<len(cart_foods);i++{
					left_stock,_ := strconv.Atoi(multi.Get("food:"+food_ids[i]+":stock").Val())
					if(food_counts[i]>left_stock){
						discarded=true
						multi.Discard()
						break
					}else{
						food_stock[i]=left_stock
					}
				}
				if discarded {
					break;
				}
			_,err:=multi.Exec(func() error{
				for i:=0;i<len(cart_foods);i++{
					multi.Set("food:"+food_ids[i]+":stock",strconv.Itoa(food_stock[i]-food_counts[i]),0)
				}
				return nil
			})
			if err==nil {
				ok = true
			}
		}
		
		if discarded{
			Response(w,403,Reply{"FOOD_OUT_OF_STOCK","食物库存不足"})
			return
		}
		
		if ok{
			client.Set(user_id+":order",t.CartId,0)
			for i:=0;i<len(cart_foods);i++{
				id,_ := strconv.Atoi(food_ids[i])
				foods.update(id,food_stock[i]-food_counts[i])
			}
			Response(w,200,OrderPostReply{t.CartId})
		}
		
		
	}else if r.Method=="GET"{
		r.ParseForm();
		
		
		order_id,err := client.Get(user_id+":order").Result()
		if err!=nil{
			Response(w,200,EmptyReply{})
			return 
		}
		
		cart_foods:=client.LRange(order_id+":cart_foods",0,2).Val()
		
		total := 0 
		var replyItem OrderGetReplyItem
		replyItem.Id=order_id
		replyItem.Items=make([]interface{},len(cart_foods))
		for i:=0;i<len(cart_foods);i++{
			strs := strings.FieldsFunc(cart_foods[i],func(s rune) bool{
				return s==':'
			})
			food_id,_ := strconv.Atoi(strs[0])
			count,_ := strconv.Atoi(strs[1])
			data := make(map[string]int)
			data["food_id"] = food_id
			data["count"] = count
			replyItem.Items[i] =data
			total+= count*foods.get_price(food_id)
		}
		replyItem.Total=total
		var reply [1]OrderGetReplyItem
		reply[0]=replyItem
		Response(w,200,reply)
		
	}
}