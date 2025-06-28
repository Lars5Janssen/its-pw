package pages

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/Lars5Janssen/its-pw/internal/repository"
	"github.com/Lars5Janssen/its-pw/login"
	"github.com/Lars5Janssen/its-pw/passkey"
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
	ctx  context.Context
	conn *pgx.Conn
)

// FOR DB
func GetUser(username string) passkey.PasskeyUser {
	repo := repository.New(conn)
	user, err := repo.GetUserByName(ctx, username)
	util.EasyCheck(err, "ERROR in GetUser util while getting user by name:", "err")

	u := &passkey.User{
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
