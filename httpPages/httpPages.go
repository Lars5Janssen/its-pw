package pages

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/Lars5Janssen/its-pw/internal/repository"
	"github.com/Lars5Janssen/its-pw/login"
	"github.com/Lars5Janssen/its-pw/passkey"
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

var (
	webAuthn  *webauthn.WebAuthn
	err       error
	datastore passkey.PasskeyStore
	l         log.Logger

	// DB
	ctx  context.Context
	conn *pgx.Conn
)

func InitPasskeys(logger log.Logger, context context.Context, connection *pgx.Conn) {
	ctx = context
	conn = connection

	l = logger
	datastore = passkey.NewInMem(l)
	wconfig := &webauthn.Config{
		RPDisplayName: "ITS123",
		// RPID:          "crisp-kangaroo-modern.ngrok-free.app",
		RPID: "localhost",
		RPOrigins: []string{
			"https://crisp-kangaroo-modern.ngrok-free.app",
			"localhost",
			"localhost:8080",
		},
	}

	if webAuthn, err = webauthn.New(wconfig); err != nil {
		log.Fatalln(err)
	}
}

func BeginRegistration(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("BeginRegistration\n")
	username, err := getUsername(r)
	if err != nil {
		log.Fatalf("Could not get username: %s\n", err.Error())
	}
	repo := repository.New(conn)
	repo.CreateUser(ctx, repository.CreateUserParams{
		ID:          []byte(uuid.NewString()),
		Name:        username,
		DisplayName: username,
	})
	user := datastore.GetOrCreateUser(username)
	options, session, err := webAuthn.BeginRegistration(user)
	if err != nil {
		msg := fmt.Sprintf("can't begin registration: %s", err.Error())
		l.Printf("ERROR %s", msg)
		JSONResponse(w, msg, http.StatusBadRequest)
		return
	}

	t, err := datastore.GenSessionID()
	if err != nil {
		l.Printf("ERROR can't gen session id: %s", err.Error())
		panic(err)
	}

	datastore.SaveSession(t, *session)
	SessionUserMap[t] = username

	http.SetCookie(w, &http.Cookie{
		Name:     "sid",
		Value:    t,
		Path:     "/app/beginRegistration",
		MaxAge:   3600,
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	http.SetCookie(w, &http.Cookie{
		Name:  "sidfix",
		Value: t,
	})

	JSONResponse(w, options, http.StatusOK)

}

func EndRegistration(w http.ResponseWriter, r *http.Request) {
	l.Println("END Registration")
	util.PrintMap(SessionUserMap)
	sid, err := r.Cookie("sid")
	if err != nil {
		l.Printf("ERROR cant get sid: %s", err.Error())
		panic(err)
	}
	sidfix, err := r.Cookie("sidfix")
	if err != nil {
		l.Printf("ERROR cant get sidfix: %s", err.Error())
		panic(err)
	}

	session, _ := datastore.GetSession(sidfix.Value)
	username, b := SessionUserMap[string(session.UserID)]
	fmt.Println("sessionMap")
	fmt.Println(b)
	util.PrintMap(SessionUserMap)
	fmt.Println("session")
	fmt.Println(string(session.UserID))
	fmt.Println("username")
	fmt.Println(username)
	user := datastore.GetOrCreateUser(username)

	credential, err := webAuthn.FinishRegistration(user, session, r)
	if err != nil {
		msg := fmt.Sprintf("can't finish registration: %s", err.Error())
		l.Printf("ERROR: %s", msg)
		http.SetCookie(w, &http.Cookie{
			Name:  "sid",
			Value: "",
		})
		JSONResponse(w, msg, http.StatusBadRequest)
		return
	}
	user.AddCredential(credential)
	datastore.SaveUser(user)
	datastore.DeleteSession(sid.Value)
	http.SetCookie(w, &http.Cookie{
		Name:  "sid",
		Value: "",
	})

	l.Printf("INFO passkey reg finished")
	JSONResponse(w, "Reg Succsess", http.StatusOK)
}

func JSONResponse(w http.ResponseWriter, data interface{}, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

// getUsername is a helper function to extract the username from json request
func getUsername(r *http.Request) (string, error) {
	type Username struct {
		Username string `json:"username"`
	}
	var u Username
	if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
		return "", err
	}

	return u.Username, nil
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

func LoginPage(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
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

		sessionToken := uuid.NewString()
		expiresAt := time.Now().Add(120 * time.Second)

		login.AddSession(sessionToken, username, expiresAt)
		http.SetCookie(w, &http.Cookie{
			Name:    "session_token",
			Value:   sessionToken,
			Expires: expiresAt,
		})

		fmt.Printf("User %s logged in\n", username)
		http.Redirect(w, r, "/app/welcome", http.StatusSeeOther)
		return
	}
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
