package api

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"runtime/debug"
	"time"
	"runtime"
)

type Reply struct {
	Code    string `json:"code"`
	Message string `json:"message"`
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

func Parser(body []byte, value interface{}) error {
	err := json.Unmarshal(body, value)
	if err != nil {
		//	log.Println("parser error")
	} else {
		//	log.Println("parser success")
	}
	return err
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

func RandStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func Error(w http.ResponseWriter, err error, message string) {
	log.Println(message)
	http.Error(w, err.Error(), http.StatusInternalServerError)
}

func Init() {
	fmt.Println("init environment")
	//To ensure every server will have same random strings in initialization.
	rand.Seed(2)
	//Disable default GC.
	debug.SetGCPercent(-1)
	
	go func(){
		//Run GC every 10 Minutes. Reduce the frequency of GC to improve performance.
		time.Sleep(10 * time.Minute)
		runtime.GC()
	}()


	InitNewMysqlClient()
	InitNewRedisClient()

	InitFood()
	InitUser()
	//To ensure the random strings are different among servers.
	rand.Seed(time.Now().UnixNano())
}
