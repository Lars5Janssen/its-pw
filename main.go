package main

import (
	"bufio"
	"crypto/sha256"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"gopkg.in/yaml.v3"
)

func main() {
	http.HandleFunc("/app/", LoginPage)
	http.HandleFunc("/app/login", LoginPage)
	http.HandleFunc("/app/welcome", WelcomePage)
	http.HandleFunc("/app/signup", SignUp)

	// http.HandleFunc("/", LoginPage)
	// http.HandleFunc("/login", LoginPage)
	// http.HandleFunc("/welcome", WelcomePage)
	// http.HandleFunc("/signup", SignUp)

	fmt.Println("Server started")
	http.ListenAndServe(":8080", nil)

}

func WelcomePage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Welcome, you are logged in!")
}

func LoginPage(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		username := r.FormValue("username")
		password := r.FormValue("password")

		result := checkLogin(username, password)
		if result {
			fmt.Printf("User %s logged in\n", username)
			http.Redirect(w, r, "/app/welcome", http.StatusSeeOther)
			// http.Redirect(w, r, "/welcome", http.StatusSeeOther)
			return
		}
		fmt.Printf("Invalid credentials entered\n")
		fmt.Fprintf(w, "Invalid credentials")
		return
	}

	tmpl, err := template.ParseFiles("templates/landing.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, nil)

}

func SignUp(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		username := r.FormValue("username")
		password := r.FormValue("password")

		config := readYaml("config.yaml")
		credentials_file, exists_file := config["credentials_file"]
		if !exists_file {
			log.Fatal("cred file not found?")
		}
		cred_path := credentials_file
		f, os_err := os.Create(cred_path)
		check(os_err)
		creds := readCreds("config.yaml")
		creds[username] = hashMe(password)
		new_creds, cred_err := yaml.Marshal(&creds)
		check(cred_err)
		defer f.Close()
		writer := bufio.NewWriter(f)
		_, w_err := writer.WriteString(string(new_creds))
		check(w_err)
		writer.Flush()

		fmt.Printf("New Login for %s registered\n", username)

		http.Redirect(w, r, "/app/login", http.StatusSeeOther)
		// http.Redirect(w, r, "/login", http.StatusSeeOther)

		return
	}
}

func check(e error) {
	if e != nil {
		log.Fatal(e)
	}
}

func hashMe(toHash string) string {
	h := sha256.New()
	h.Write([]byte(toHash))
	return string(h.Sum(nil))
}

func checkLogin(username string, password string) bool {

	creds := readCreds("config.yaml")

	found_password, user_exists := creds[username]
	if !user_exists {
		return false
	}
	if found_password == hashMe(password) {
		return true
	}
	return false
}

func readYaml(path string) map[string]string {
	file, err := ioutil.ReadFile(path)
	check(err)
	filemap := make(map[interface{}]interface{})
	err2 := yaml.Unmarshal(file, &filemap)
	check(err2)
	finalmap := make(map[string]string)
	for k, v := range filemap {
		key := fmt.Sprintf("%s", k)
		value := fmt.Sprintf("%s", v)
		finalmap[key] = value
	}
	return finalmap
}

func readCreds(config_path string) map[string]string {
	config := readYaml(config_path)
	credentials_store_method, exists := config["credentials_store_method"]
	if !exists {
		log.Fatal("credentials_store_method does not exist")
	}
	if credentials_store_method != "yaml" {
		log.Fatal("not yet implemented")
	}
	credentials_file, exists_file := config["credentials_file"]
	if !exists_file {
		log.Fatal("cred file not found?")
	}
	cred_path := credentials_file
	creds_to_convert := readYaml(cred_path)
	creds := make(map[string]string)
	for k, v := range creds_to_convert {
		key := fmt.Sprintf("%s", k)
		value := fmt.Sprintf("%s", v)
		creds[key] = value
	}
	return creds

}
