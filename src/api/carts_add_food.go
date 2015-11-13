package api
import (
	"strconv"
	"net/http"
	"strings"
	"io/ioutil"
)

type EmptyReply struct{
	
}


type CartsAddArgs struct{
	FoodId	int `json:"food_id"`
	Count	int `json:"count"`
}

func GetCartidFromUrl(urlpath string) string{
	return strings.FieldsFunc(urlpath,func(c rune)bool{
		return c=='/'
	})[1]
}

func CartsAddFoodHandler(w http.ResponseWriter, r *http.Request){
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
	cart_id := GetCartidFromUrl(r.URL.Path)
	if !client.SIsMember("ALL_CARTS",cart_id).Val(){
		Response(w,404,Reply{"CART_NOT_FOUND","篮子不存在"})
		return
	}
	
	user_id :=strconv.Itoa(user.id)
	user_cart_id := client.Get(user_id+":carts").Val()
	if cart_id != user_cart_id{
		Response(w,401,Reply{"NOT_AUTHORIZED_TO_ACCESS_CART","无权限访问指定的篮子"})
		return
	}
	
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		Error(w, err, "can not read body")
		return
	}
	if len(body) == 0 {
		Response(w, 400, Reply{"EMPTY_REQUEST", "请求体为空"})
		return
	}
	
	var t CartsAddArgs
	err = Parser(body, &t)

	if err!=nil {
		Response(w, 400, Reply{"MALFORMED_JSON", "格式错误"})
		return
	}

	
	if !foods.check(t.FoodId) {
		Response(w,404,Reply{"FOOD_NOT_FOUND","食物不存在"})
		return
	}

	foods := client.LRange(cart_id+":cart_foods",0,2).Val();
	total_foods:=0
	for i:=0;i<len(foods);i++{
		strs := strings.FieldsFunc(foods[i],func(s rune) bool{
					return s==':'
		})
		food_count,_:= strconv.Atoi(strs[1])
		total_foods+=food_count
	}
	if total_foods+t.Count>3{
		Response(w,403,Reply{"FOOD_OUT_OF_LIMIT","篮子中食物数量超过了三个"})
		return
	}
	
	client.LPush(cart_id+":cart_foods",strconv.Itoa(t.FoodId)+":"+strconv.Itoa(t.Count))
	//Response(w,200,CartsReply{cart_id})
	Response(w,204,EmptyReply{})
}