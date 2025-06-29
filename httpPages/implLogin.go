package pages

import (
	"encoding/json"
	"net/http"

	"github.com/Lars5Janssen/its-pw/util"
)

func SendLogin(w http.ResponseWriter, r *http.Request) {
	l.Println("SEND LOGIN")
	data, err := getSendLoginData(r)
	util.EasyCheck(err, "ERROR in SendLogin during getSendLoginData: ", "err")
	l.Println(data)
}

type implSendLoginData struct {
	Username      string `json:"username"`
	EncryptedData string `json:"encryptedData"`
}

func getSendLoginData(r *http.Request) (implSendLoginData, error) {
	var data implSendLoginData
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		return implSendLoginData{}, err
	}
	return data, nil
}

type implAwnserLoginData struct {
	EncryptedData string `json:"encryptedData"`
}

type implMessageData struct {
	Sid           string `json:"sid"`
	EncryptedData string `json:"encryptedData"`
}
