// Go hello-world implementation for eleme/hackathon.

package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"sync"
	//	"sync/atomic"
	//	"strconv"
	"runtime"
	_ "api"
)

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func RandStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func getDb() (*sql.DB, error) {
	hostname := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	dbname := os.Getenv("DB_NAME")
	username := os.Getenv("DB_USER")
	pwd := os.Getenv("DB_PASS")

	str := username + ":" + pwd + "@tcp(" + hostname + ":" + port + ")/" + dbname
	log.Println(str)
	db, error := sql.Open("mysql", str)
	if error != nil {
		return db, error
	}
	error = db.Ping()
	return db, error
}

type loginData struct {
	Username string `json:"username"`
	Password string `json:"password"`
}
type Reply struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}
type loginReply struct {
	Userid   int    `json:"user_id"`
	Username string `json:"username"`
	Token    string `json:"access_token"`
}

var db *sql.DB

func Error(w http.ResponseWriter, err error, message string) {
	log.Println(message)
	http.Error(w, err.Error(), http.StatusInternalServerError)
}

func Parser(body []byte, value interface{}) error {
	err := json.Unmarshal(body, value)
	if err != nil {
		log.Println("parser error")
	} else {
		log.Println("parser success")
	}
	return err
}

func Response(w http.ResponseWriter, status int, reply interface{}) {
	w.WriteHeader(status)
	js, err := json.Marshal(reply)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
	//	log.Println(string(js))
	return
}

type Token struct {
	mu  sync.Mutex
	set map[string]bool
}

func (token *Token) add(name string) {
	token.mu.Lock()
	token.set[name] = true
	defer token.mu.Unlock()
}
func (token *Token) remove(name string) {
	token.mu.Lock()
	token.set[name] = false
	defer token.mu.Unlock()
}
func (token *Token) check(tok1 string, tok2 string) bool {
	res, ok := token.set[tok1]
	if ok == true && res == true {
		return true
	}
	res, ok = token.set[tok2]
	if ok == true && res == true {
		return true
	}
	return false
}

func check_access(token *Token, r *http.Request) bool {
	tok1 := r.Form.Get("access_token")
	tok2 := r.Header.Get("Access-Token")
	if token.check(tok1, tok2) == true {
		return true
	} else {
		return false
	}
}

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

func main() {
	runtime.GOMAXPROCS(4)
	db, err := getDb()
	if err != nil {
		fmt.Println("err: ", err)
		panic("get mysql failed")
	}
	err = db.Ping()
	if err != nil {
		fmt.Println("err: ", err)
		panic("can not ping mysql")
	}

	token := new(Token)
	token.set = make(map[string]bool)

	foods := new(Foods)
	foods.init()
	rows, err := db.Query("select * from food")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		var id, price, stock int
		rows.Scan(&id, &price, &stock)
		foods.add(id, price, stock)
	}

	host := os.Getenv("APP_HOST")
	port := os.Getenv("APP_PORT")
	if host == "" {
		host = "localhost"
	}
	if port == "" {
		port = "8080"
	}

	addr := fmt.Sprintf("%s:%s", host, port)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hello world!"))
	})

	http.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
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
			rows, err := db.Query("select id, password from user where name=?", t.Username)
			if err != nil {
				log.Fatal(err)
			}
			var pwd string
			var id int
			rows.Next()
			err = rows.Scan(&id, &pwd)
			if err != nil {
				log.Fatal(err)
			}
			defer rows.Close()

			if pwd == t.Password {
				str := RandStringRunes(40)
				log.Println(str)
				token.add(str)
				Response(w, 200, loginReply{id, t.Username, str})
			} else {
				Response(w, 403, Reply{"USER_AUTH_FAIL", "用户名或密码错误"})
			}
			return
		} else {
			Response(w, 400, Reply{"MALFORMED_JSON", "格式错误"})
			return
		}
	})

	http.HandleFunc("/foods", func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		if check_access(token, r) {
			Response(w, 200, foods.Foods[1:101])
		} else {
			Response(w, 200, "no valid access token")
		}
	})

	http.HandleFunc("/carts", func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		if check_access(token, r) {
			cart_id := RandStringRunes(40)
			log.Println(cart_id)

		} else {
			Error(w, nil, "no valid access token")
		}
	})
	http.ListenAndServe(addr, nil)
}
