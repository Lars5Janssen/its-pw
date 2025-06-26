package main

import (
	"bufio"
	"crypto/sha256"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/pquerna/otp/totp"
	"gopkg.in/yaml.v3"

	"github.com/Lars5Janssen/its-pw/httpPages"
)

func main() {
	http.HandleFunc("/app/", pages.LoginPage)
	http.HandleFunc("/app/login", pages.LoginPage)
	http.HandleFunc("/app/welcome", pages.WelcomePage)
	http.HandleFunc("/app/signup", pages.SignUp)

	fmt.Println("Server started")
	log.Fatal(http.ListenAndServe(":8080", nil))

}

type session struct {
	username string
	expiry   time.Time
}

var sessions = map[string]session{}

func (s session) isExpired() bool {
	return s.expiry.Before(time.Now())
}

var totpmap = map[string]string{}
