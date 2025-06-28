package util

import (
	"encoding/json"
	"log"
	"net/http"
)

var l log.Logger

func Init(logger log.Logger) {
	l = logger
}

func JSONResponse(w http.ResponseWriter, data interface{}, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	err := json.NewEncoder(w).Encode(data)
	EasyCheck(err, "ERROR in JSONResponse:", "err")
}

func Check(e error) {
	if e != nil {
		l.Panicln("ERROR in Check():", e.Error())
		panic(e)
	}
}

func EasyCheck(e error, v ...any) {
	if e != nil {
		for i, s := range v {
			if s == "err" {
				v[i] = e.Error()
			}
		}
		l.Println("EASYCHECK", v)
	}
}
