package pages

import (
	"context"
	"crypto/rand"
	"encoding/base64"
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
	webAuthn *webauthn.WebAuthn
	err      error
	l        log.Logger

	// DB
	ctx  context.Context
	conn *pgx.Conn
)

func InitPasskeys(logger log.Logger, context context.Context, connection *pgx.Conn) {
	ctx = context
	conn = connection

	l = logger
	wconfig := &webauthn.Config{
		RPDisplayName: "ITS123",
		// RPID:          "crisp-kangaroo-modern.ngrok-free.app",
		RPID: "localhost",
		RPOrigins: []string{
			"https://crisp-kangaroo-modern.ngrok-free.app",
			"localhost",
			"http://localhost:8080",
			"localhost:8080",
		},
	}

	if webAuthn, err = webauthn.New(wconfig); err != nil {
		log.Fatalln(err)
	}
}

func BeginRegistration(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("BeginRegistration\n")
	repo := repository.New(conn)

	username, err := getUsername(r)

	if err != nil {
		log.Fatalf("Could not get username: %s\n", err.Error())
	}

	repo.CreateUser(ctx, repository.CreateUserParams{
		ID:          []byte(uuid.NewString()),
		Name:        username,
		DisplayName: username,
	})

	user := GetUser(username)
	options, session, err := webAuthn.BeginRegistration(user)

	if err != nil {
		msg := fmt.Sprintf("can't begin registration: %s", err.Error())
		l.Printf("ERROR %s", msg)
		JSONResponse(w, msg, http.StatusBadRequest)
		return
	}

	t, err := GenSessionID()
	if err != nil {
		l.Printf("ERROR can't gen session id: %s", err.Error())
		panic(err)
	}

	repoUser, _ := repo.GetUserByName(ctx, username)
	sessionData, _ := json.Marshal(session)

	repo.CreateSession(ctx, repository.CreateSessionParams{
		UserID:      repoUser.ID,
		SessionID:   t,
		SessionData: sessionData,
	})

	// http.SetCookie(w, &http.Cookie{
	// 	Name:     "sid",
	// 	Value:    t,
	// 	Path:     "/app/beginRegistration",
	// 	MaxAge:   3600,
	// 	Secure:   true,
	// 	HttpOnly: true,
	// 	SameSite: http.SameSiteLaxMode,
	// })
	http.SetCookie(w, &http.Cookie{
		Name:  "sidfix",
		Value: t,
	})
	http.SetCookie(w, &http.Cookie{
		Name:  "name",
		Value: username,
	})

	JSONResponse(w, options, http.StatusOK)
}

func GenSessionID() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	return base64.URLEncoding.EncodeToString(b), nil
}

func EndRegistration(w http.ResponseWriter, r *http.Request) {
	l.Println("END Registration")
	sidfix, err := r.Cookie("sidfix")
	if err != nil {
		l.Printf("ERROR cant get sidfix: %s", err.Error())
		panic(err)
	}
	name, err := r.Cookie("name")
	if err != nil {
		l.Printf("ERROR cant get name: %s", err.Error())
		panic(err)
	}

	repo := repository.New(conn)
	user, _ := repo.GetUserByName(ctx, name.Value)
	marshaledSession, _ := repo.GetSessionBySessionId(ctx, sidfix.Value)
	puser := GetUser(name.Value)
	var session webauthn.SessionData
	json.Unmarshal(marshaledSession.SessionData, &session)

	credential, err := webAuthn.FinishRegistration(puser, session, r)
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

	jsonCreds, _ := json.Marshal(credential)
	repo.UpdateUserCredentials(ctx, repository.UpdateUserCredentialsParams{
		ID:          user.ID,
		Credentials: jsonCreds,
	})
	repo.DeleteSessionByUserId(ctx, user.ID)
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
