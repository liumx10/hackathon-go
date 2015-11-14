package api

import (
	"strconv"
	"net/http"
)
type CartsReply struct {
	CartId   string    `json:"cart_id"`
}
func CartsHandler(w http.ResponseWriter, r *http.Request){
	err := r.ParseForm();
	if err!=nil {
		Response(w, 400, Reply{"MALFORMED_JSON", "格式错误"})
		return
	}

	user,err := users.get_user_by_request(r)
	if err!= nil{
		Response(w, 401, Reply{"INVALID_ACCESS_TOKEN", "无效的令牌"})
		return
	}
	cart_id := RandStringRunes(32)
	client.SAdd("ALL_CARTS",cart_id)
	client.Set(strconv.Itoa(user.id)+":carts",cart_id,0)
	
	Response(w,200,CartsReply{cart_id})
}
