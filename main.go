package main

import (
	"fmt"
	"html/template"
	"net/http"
)

func main() {
	http.HandleFunc("/", LoginPage)
	http.HandleFunc("/login", LoginPage)
	http.HandleFunc("/welcome", WelcomePage)

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	fmt.Println("Server started")
	http.ListenAndServe(":8080", nil)

}

func LoginPage(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		username := r.FormValue("username")
		password := r.FormValue("password")

		result := checkLogin(username, password)
		if result {
			http.Redirect(w, r, "/welcome", http.StatusSeeOther)
			return
		}
		fmt.Fprintf(w, "Invalid credentials")
		return
	}

	tmpl, err := template.ParseFiles("templates/login.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, nil)

}

func checkLogin(username string, password string) bool {
	if username == password {
		return true
	}
	return false
}

func WelcomePage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Welcome, you are logged in!")
}
