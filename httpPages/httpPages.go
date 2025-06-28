package pages

import (
	"fmt"
	"html/template"
	"net/http"
	"os"

	"github.com/Lars5Janssen/its-pw/login"
	"github.com/Lars5Janssen/its-pw/util"
)

func WelcomePage(w http.ResponseWriter, r *http.Request) {
	vaild, status, _, username := login.CheckSessionToken(w, r)
	if !vaild {
		fmt.Println("Login Page was accsessed with invalid Token")
		http.Redirect(w, r, "/app/login", status)
		return
	}
	fmt.Println("Login Page was accsessed with valid Token")
	userString := fmt.Sprintf("Welcome, %s\nyou are logged in!", username)
	fmt.Fprint(w, userString)
}

func SignUp(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		username := r.FormValue("username")
		password := r.FormValue("password")
		fmt.Printf("SignUp Endpoint was accsessed for %s with pw:\"%s\"\n", username, password)

		login.AddUser(username, password)
		login.GenerateTOTP(username)

		fmt.Printf("New Login for %s registered\n", username)

		http.Redirect(w, r, "/app/login", http.StatusSeeOther)
		return
	}
}

func LandingPage(w http.ResponseWriter, r* http.Request) {
	fmt.Println("LoginPage was accsessed")

	html, err := os.ReadFile("templates/landing.html")
	util.Check(err)
	tmpl, err := template.New("base").Parse(string(html))
	util.Check(err)

	script, err := os.ReadFile("templates/script.js")
	util.Check(err)
	otherscript, err := os.ReadFile("templates/index.es5.umd.min.js")
	util.Check(err)
	err = tmpl.ExecuteTemplate(
		w,
		"base",
		template.JS(string(otherscript)+"\n"+string(script)),
	)
}

