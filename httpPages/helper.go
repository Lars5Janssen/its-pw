package pages

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/Lars5Janssen/its-pw/internal/repository"
	"github.com/Lars5Janssen/its-pw/login"
	"github.com/Lars5Janssen/its-pw/util"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

var (
	webAuthn *webauthn.WebAuthn
	err      error
	l        log.Logger

	// DB
	ctx      context.Context
	conn     *pgx.Conn
	globalID string
)

// FOR DB
func GetUser(username string) PasskeyUser {
	repo := repository.New(conn)
	user, err := repo.GetUserByName(ctx, username)
	util.EasyCheck(err, "ERROR in GetUser util while getting user by name:", "err")

	u := &User{
		ID:          user.ID,
		Name:        user.Name,
		DisplayName: user.DisplayName,
	}
	var creds *webauthn.Credential
	err = json.Unmarshal(user.Credentials, &creds)
	if err == nil {
		u.AddCredential(creds)
	}

	return u
}

func LocationTest(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, globalID)
	return
}

func locationTest() bool {

	// url := "https://crisp-kangaroo-modern.ngrok-free.app/app/LocationTest"
	url := "localhost/app/LocationTest"
	resp, err := http.Get(url)
	if err != nil {
		l.Println("Location Test Result: Local, URL not reachable")
		return false
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		l.Println("Location Test Result: Local, Body Error:", err.Error())
		return false
	}
	if string(data) == globalID {
		l.Println("Location Test Result: Public")
		return true
	}
	l.Println("Location Test Data: ", string(data), "\ngloablID: ", globalID)
	l.Println("Location Test Result: Local")
	return false
}

func InitPasskeys(logger log.Logger, context context.Context, connection *pgx.Conn) {
	ctx = context
	conn = connection
	globalID = uuid.NewString()

	l = logger
	rpid := "localhost"
	if locationTest() {
		rpid = "crisp-kangaroo-modern.ngrok-free.app"
	}
	wconfig := &webauthn.Config{
		RPDisplayName: "ITS123",
		RPID:          rpid,
		RPOrigins: []string{
			"https://crisp-kangaroo-modern.ngrok-free.app",
			"localhost",
			"https://localhost",
			"http://localhost:8080",
			"http://localhost:8765",
			"https://localhost:8080",
			"localhost:8080",
		},
	}

	if webAuthn, err = webauthn.New(wconfig); err != nil {
		log.Fatalln(err)
	}
}

func setLoginSessionToken(w http.ResponseWriter, username string) {
	sessionToken := uuid.NewString()
	expiresAt := time.Now().Add(120 * time.Second)

	login.AddSession(sessionToken, username, expiresAt)
	http.SetCookie(w, &http.Cookie{
		Name:    "session_token",
		Value:   sessionToken,
		Expires: expiresAt,
	})
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

func GenSessionID() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	return base64.URLEncoding.EncodeToString(b), nil
}
