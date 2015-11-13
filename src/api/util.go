package api

import (
	"net/http"
	"encoding/json"
	"math/rand"
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

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

func RandStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}


