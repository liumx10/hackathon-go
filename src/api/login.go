package api

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

type User struct {
	id   int
	name string
	pwd  string
}
type Users struct {
	Users map[string]User
}

func (users *Users) init() {
	users.Users = make(map[string]User)
}
func (users *Users) add(id int, name string, pwd string) {
	users.Users[name] = User{id, name, pwd}
}
func (users *Users) check(name string, pwd string) bool {
	v, ok := users.Users[name]
	if ok == false {
		return false
	} else if v.pwd == pwd {
		return true
	}
	return false
}
func (users *Users) getuser(name string) User {
	v, ok := users.Users[name]
	if !ok {
		return User{-1, "", ""}
	}
	return v
}

var users *Users

func InitUser() {
	users = new(Users)
	users.init()
	rows, err := db.Query("select * from user")
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
	defer rows.Close()
	for rows.Next() {
		var id int
		var name, pwd string
		rows.Scan(&id, &name, &pwd)
		users.add(id, name, pwd)
	}
	i := 0
	for _, user := range users.Users {
		fmt.Println(user.id, user.name, user.pwd)
		i++
		if i > 10 {
			break
		}
	}
}

type loginData struct {
	Username string `json:"username"`
	Password string `json:"password"`
}
type loginReply struct {
	Userid   int    `json:"user_id"`
	Username string `json:"username"`
	Token    string `json:"access_token"`
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		Error(w, err, "can not read body")
		return
	}
	if len(body) == 0 {
		Response(w, 400, Reply{"EMPTY_REQUEST", "请求体为空"})
	}
	var t loginData
	err = Parser(body, &t)

	if err == nil {
		user := users.getuser(t.Username)
		if user.id != -1 && user.pwd == t.Password {
			str := RandStringRunes(40)
			log.Println(str)
			err := client.Set(t.Username, str, 0).Err()
			if err != nil {
				fmt.Println("redis add token failed")
				Response(w, 500, "")
			}
			Response(w, 200, loginReply{user.id, t.Username, str})
		} else {
			Response(w, 403, Reply{"USER_AUTH_FAIL", "用户名或密码错误"})
		}
		return
	} else {
		Response(w, 400, Reply{"MALFORMED_JSON", "格式错误"})
		return
	}
}
