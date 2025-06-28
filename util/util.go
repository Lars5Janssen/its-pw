package util

import (
	"encoding/json"
	"log"
	"net/http"
)

var l log.Logger

func Init(l log.Logger) {
	l = l
}

func JSONResponse(w http.ResponseWriter, data interface{}, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func Check(e error) {
	if e != nil {
		l.Panicln("ERROR in Check():", e.Error())
		panic(e)
	}
}

func EasyCheck(e error, v ...any) {
	if e != nil {
		l.Println(v)
	}
}
