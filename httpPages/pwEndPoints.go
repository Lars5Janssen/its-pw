package pages

import (
	"fmt"
	"net/http"

	"github.com/Lars5Janssen/its-pw/login"
)

func Login(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	password := r.FormValue("password")
	totpCode := r.FormValue("totp")
	fmt.Printf(
		"Login Endpoint was accsessed for %s with pw:\"%s\" and totp:%s\n",
		username,
		password,
		totpCode,
	)

	result := login.CheckLogin(username, password, totpCode)
	if !result {
		fmt.Printf("Invalid credentials entered\n")
		fmt.Fprintf(w, "Invalid credentials")
		return
	}

	setLoginSessionToken(w, username)

	l.Printf("User %s logged in\n", username)
	http.Redirect(w, r, "/app/welcome", http.StatusSeeOther)
}

func SignUp(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	password := r.FormValue("password")
	fmt.Printf("SignUp Endpoint was accsessed for %s with pw:\"%s\"\n", username, password)

	login.AddUser(username, password)
	login.GenerateTOTP(username)

	fmt.Printf("New Login for %s registered\n", username)

	http.Redirect(w, r, "/app/login", http.StatusSeeOther)
}
