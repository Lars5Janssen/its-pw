package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/Lars5Janssen/its-pw/files"
	"github.com/Lars5Janssen/its-pw/httpPages"
	"github.com/Lars5Janssen/its-pw/login"
)

var l log.Logger

func main() {
	l = *log.Default()
	http.HandleFunc("/app/", pages.LoginPage)
	http.HandleFunc("/app/login", pages.LoginPage)
	http.HandleFunc("/app/welcome", pages.WelcomePage)
	http.HandleFunc("/app/signup", pages.SignUp)

	http.HandleFunc("/app/beginRegistration", pages.BeginRegistration)
	http.HandleFunc("/app/endRegistration", pages.EndRegistration)

	if _, err := os.Stat("creds.yaml"); err != nil {
		files.WriteYAML("cred.yaml", make(map[string]string))
	}
	pages.InitPasskeys(l)
	login.AddDefaultUser()

	fmt.Println("Server started")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
