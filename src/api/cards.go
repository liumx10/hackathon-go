package api

import (
	"net/http"
)
type CardsReply struct {
	CardId   string    `json:"card_id "`
}
func CardsHandler(w http.ResponseWriter, r *http.Request){
	err := r.ParseForm();
	if err!=nil {
		Response(w, 400, Reply{"MALFORMED_JSON", "格式错误"})
		return
	}
	//token := r.FormValue("access_token")
	//user_id,err := check_token(token)
	user_id := "Hi"
	if err!= nil{
		Response(w, 401, Reply{"INVALID_ACCESS_TOKEN", "无效的令牌"})
	}
	card_id := RandStringRunes(32)
	go func(){
		client.Set(user_id+":cards",card_id,0)
	}()
	
	Response(w,200,CardsReply{card_id})
}
