package api

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"sync"
)

type User struct {
	id   int
	name string
	pwd  string
}
type Users struct {
	Users      map[string]User
	Token2User map[string]User
	m          *sync.RWMutex
}

func (users *Users) init() {
	users.Users = make(map[string]User)
	users.Token2User = make(map[string]User)
	users.m = new(sync.RWMutex)
}
func (users *Users) add(id int, name string, pwd string) {
	users.Users[name] = User{id, name, pwd}
}
func (users *Users) addtoken(tok string, user User) {
	users.Token2User[tok] = user
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

func (users *Users) get_user_by_request(r *http.Request) (User, error) {

	var err error
	tok1 := r.Form.Get("access_token")
	tok2 := r.Header.Get("Access-Token")
	client := BorrowClient()
	defer ReturnClient(client)

	var tok string
	if len(tok1) > 0 {
		tok = tok1
	}
	if len(tok2) > 0 {
		tok = tok2
	}
	if len(tok) > 0 {
		users.m.RLock()
		v, ok := users.Token2User[tok]
		users.m.RUnlock()
		if ok {
			fmt.Println("token cached")
			return v, nil
		} else {
			name, _ := client.Get("token2name:" + tok).Result()
			user := users.getuser(name)
			if user.id > 0 {
				users.m.Lock()
				users.Token2User[tok] = user
				users.m.Unlock()
				return user, nil
			}
			err = errors.New("Invalid token!")
			return User{}, err
		}
	}
	err = errors.New("No token")
	return User{}, err
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
		return
	}
	var t loginData
	err = Parser(body, &t)

	if err == nil {
		user := users.getuser(t.Username)
		if user.id != -1 && user.pwd == t.Password {
			str := RandStringRunes(40)
			//log.Println(str)
			//log.Println(t.Username)

			client := BorrowClient()
			defer ReturnClient(client)
			pipeline := client.Pipeline()
			defer pipeline.Close()
			//		pipeline.Set("name2token:"+t.Username, str, 0).Err()
			pipeline.Set("token2name:"+str, t.Username, 0).Err()
			pipeline.Exec()

			users.m.Lock()
			users.addtoken(str, user)
			users.m.Unlock()

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
