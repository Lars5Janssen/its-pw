package login

import (
	"crypto/sha256"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/pquerna/otp/totp"

	"github.com/Lars5Janssen/its-pw/files"
)

var sessions = map[string]Session{}
var totpmap = map[string]string{}

type Session struct {
	username string
	expiry   time.Time
}

func (s Session) isExpired() bool {
	return s.expiry.Before(time.Now())
}

func ReadCreds() map[string]string {
	return files.ReadYaml("creds.yaml")
}

func WriteCreds(creds map[string]string) {
	files.WriteYAML("creds.yaml", creds)
}

func hashMe(toHash string) string {
	h := sha256.New()
	h.Write([]byte(toHash))
	return string(h.Sum(nil))
}

func AddSession(uuid string, username string, expiresAt time.Time) {
	sessions[uuid] = Session{
		username: username,
		expiry:   expiresAt,
	}
}

func CheckLogin(username string, password string, totpCode string) bool {

	creds := ReadCreds()

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

func AddUser(username string, password string) {
	creds := ReadCreds()
	creds[username] = hashMe(password)
}

func GenerateTOTP(username string) {
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

func CheckSessionToken(w http.ResponseWriter, r *http.Request) (bool, int, Session) {
	c, err := r.Cookie("session_token")
	if err != nil {
		if err == http.ErrNoCookie {
			w.WriteHeader(http.StatusUnauthorized)
			return false, http.StatusUnauthorized, Session{}
		}

		w.WriteHeader(http.StatusBadRequest)
		return false, http.StatusBadRequest, Session{}
	}
	sessionToken := c.Value
	userSession, exists := sessions[sessionToken]
	if !exists {
		w.WriteHeader(http.StatusUnauthorized)
		return false, http.StatusUnauthorized, Session{}
	}
	if userSession.isExpired() {
		delete(sessions, sessionToken)
		w.WriteHeader(http.StatusUnauthorized)
		return false, http.StatusUnauthorized, Session{}
	}

	return true, http.StatusOK, userSession
}
