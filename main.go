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

func main() {
	http.HandleFunc("/app/", pages.LoginPage)
	http.HandleFunc("/app/login", pages.LoginPage)
	http.HandleFunc("/app/welcome", pages.WelcomePage)
	http.HandleFunc("/app/signup", pages.SignUp)

	if _, err := os.Stat("creds.yaml"); err != nil {
		files.WriteYAML("cred.yaml", make(map[string]string))
	}
	login.AddDefaultUser()

	fmt.Println("Server started")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
