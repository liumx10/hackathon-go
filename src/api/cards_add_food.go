package api
import (
	"net/http"
	"log"
)

func CardsAddFoodHandler(w http.ResponseWriter, r *http.Request){
	err := r.ParseForm();
	if err!=nil {
		Response(w, 400, Reply{"MALFORMED_JSON", "格式错误"})
		return
	}
	//token := r.FormValue("access_token")
	//user_id,err := check_token(token)
	if err!= nil{
		Response(w, 401, Reply{"INVALID_ACCESS_TOKEN", "无效的令牌"})
	}
	//food_id := r.PostFormValue("food_id")
	//food_count := r.PostFormValue("count")
	log.Println(r.Form.Encode())
	//Response(w,200,CardsReply{card_id})
}