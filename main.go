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
	"time"

	"github.com/google/uuid"
	"github.com/pquerna/otp/totp"
	"gopkg.in/yaml.v3"
)

func main() {
	http.HandleFunc("/app/", LoginPage)
	http.HandleFunc("/app/login", LoginPage)
	http.HandleFunc("/app/welcome", WelcomePage)
	http.HandleFunc("/app/signup", SignUp)

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

func generateTOTP(username string) {
	key, keyErr := totp.Generate(totp.GenerateOpts{
		Issuer:      "localhost",
		AccountName: username,
		SecretSize:  20,
	})

	if keyErr != nil {
		log.Fatal("Something has gone wrong during otp generation")
	}

	totpmap[username] = key.Secret()
	fmt.Println(key.Secret())
}

func WelcomePage(w http.ResponseWriter, r *http.Request) {
	vaild, status, sessionToken := checkSessionToken(w, r)
	if !vaild {
		http.Redirect(w, r, "/app/login", status)
		return
	}
	userString := fmt.Sprintf("Welcome, %s\nyou are logged in!", sessionToken.username)
	fmt.Fprint(w, userString)
}
func checkSessionToken(w http.ResponseWriter, r *http.Request) (bool, int, session) {
	c, err := r.Cookie("session_token")
	if err != nil {
		if err == http.ErrNoCookie {
			w.WriteHeader(http.StatusUnauthorized)
			return false, http.StatusUnauthorized, session{}
		}

		w.WriteHeader(http.StatusBadRequest)
		return false, http.StatusBadRequest, session{}
	}
	sessionToken := c.Value
	userSession, exists := sessions[sessionToken]
	if !exists {
		w.WriteHeader(http.StatusUnauthorized)
		return false, http.StatusUnauthorized, session{}
	}
	if userSession.isExpired() {
		delete(sessions, sessionToken)
		w.WriteHeader(http.StatusUnauthorized)
		return false, http.StatusUnauthorized, session{}
	}

	return true, http.StatusOK, userSession
}

func LoginPage(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		username := r.FormValue("username")
		password := r.FormValue("password")
		totpCode := r.FormValue("totp")

		result := checkLogin(username, password, totpCode)
		if !result {
			fmt.Printf("Invalid credentials entered\n")
			fmt.Fprintf(w, "Invalid credentials")
			return
		}

		sessionToken := uuid.NewString()
		expiresAt := time.Now().Add(120 * time.Second)

		sessions[sessionToken] = session{
			username: username,
			expiry:   expiresAt,
		}

		http.SetCookie(w, &http.Cookie{
			Name:    "session_token",
			Value:   sessionToken,
			Expires: expiresAt,
		})

		fmt.Printf("User %s logged in\n", username)
		http.Redirect(w, r, "/app/welcome", http.StatusSeeOther)
		return
	}

	tmpl, err := template.ParseFiles("templates/landing.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, nil)

}

func writeYAML(path string, content interface{}) {
	f, os_err := os.Create(path)
	defer f.Close()
	check(os_err)
	new_content, cred_err := yaml.Marshal(&content)
	check(cred_err)
	writer := bufio.NewWriter(f)
	_, w_err := writer.WriteString(string(new_content))
	check(w_err)
	writer.Flush()

}

func AddUser(username string, password string) {
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

}

func SignUp(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		username := r.FormValue("username")
		password := r.FormValue("password")

		AddUser(username, password)
		generateTOTP(username)

		fmt.Printf("New Login for %s registered\n", username)

		http.Redirect(w, r, "/app/login", http.StatusSeeOther)
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

func checkLogin(username string, password string, totpCode string) bool {

	creds := readCreds("config.yaml")

	found_password, user_exists := creds[username]
	if !user_exists {
		return false
	}
	if found_password != hashMe(password) {
		return false
	}
	userSecret, exists := totpmap[username]
	if !exists {
		return false
	}
	if totp.Validate(totpCode, userSecret) {
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
