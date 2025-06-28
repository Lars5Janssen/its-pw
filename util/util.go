package util

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

func PrintMap(m map[string]string) {
	for k, v := range m {
		fmt.Printf("k:%s,v:%s\n", k, v)
	}
}

func JSONResponse(w http.ResponseWriter, data interface{}, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func ItoSmap(m map[interface{}]interface{}) map[string]string {
	stringMap := map[string]string{}
	for k, v := range m {
		key := fmt.Sprintf("%s", k)
		value := fmt.Sprintf("%s", v)
		stringMap[key] = value
	}
	return stringMap
}

func Check(e error) {
	if e != nil {
		log.Fatal(e)
	}
}
