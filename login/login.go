package login

import (
	"context"
	"crypto/sha256"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/pquerna/otp/totp"

	"github.com/Lars5Janssen/its-pw/internal/repository"
)

var (
	l    log.Logger
	ctx  context.Context
	conn *pgx.Conn
)

func InitLogin(log log.Logger, conntext context.Context, connection *pgx.Conn) {
	l = log
	ctx = conntext
	conn = connection
}

type Session struct {
	username string
	expiry   time.Time
}

func (s Session) isExpired() bool {
	return s.expiry.Before(time.Now())
}

// func ReadCreds() map[string]string {
// 	fmt.Println("Reading Credentials")
// 	return files.ReadYaml("creds.yaml")
// }
//
// func WriteCreds(creds map[string]string) {
// 	fmt.Println("Writing Credentials")
// 	files.WriteYAML("creds.yaml", creds)
// }

func hashMe(toHash string) string {
	h := sha256.New()
	h.Write([]byte(toHash))
	return string(h.Sum(nil))
}

func AddSession(uuid string, username string, expiresAt time.Time) {
	repo := repository.New(conn)
	repo.CreatePwUserSession(ctx, repository.CreatePwUserSessionParams{
		Uuid:      uuid,
		Username:  username,
		ExpiresAt: expiresAt,
	})
}

func CheckLogin(username string, password string, totpCode string) bool {

	repo := repository.New(conn)
	user, err := repo.GetPwUserByName(ctx, username)

	if err != nil {
		l.Println("CHECK LOGIN: error in db: ", err.Error())
		return false
	}

	if string(user.Pw) != hashMe(password) {
		l.Println("CHECK LOGIN: pw do not match")
		return false
	}

	if user.TotpSecret == nil || string(user.TotpSecret) == "" {
		l.Println("CHECK LOGIN: totp secret error")
		return false
	}

	if totp.Validate(totpCode, string(user.TotpSecret)) {
		l.Println("CHECK LOGIN: Check Login succsess")
		return true
	}
	l.Println("CHECK LOGIN: Check Login failure")
	return false
}

func AddDefaultUser() {
	repo := repository.New(conn)
	err := repo.DeletePwUserByName(ctx, "default")
	AddUser("default", "default")
	if err != nil {
		l.Println("ERROR in def user: ", err.Error())
	}
	totpSecret := "QACZSSNENVAXRPMVJWCY2NL6RT34W2HP"
	repo.UpdatePwUsertotpByName(ctx, repository.UpdatePwUsertotpByNameParams{
		Username:   "default",
		TotpSecret: []byte(totpSecret),
	})
}

func AddUser(username string, password string) {
	repo := repository.New(conn)
	hash := hashMe(password)
	empty := ""
	err := repo.AddPwUser(ctx, repository.AddPwUserParams{
		Username:   username,
		Pw:         []byte(hash),
		TotpSecret: []byte(empty),
	})
	if err != nil {
		l.Println("ERROR in adding user: ", username, "\n", err.Error())
	}
	l.Println("Added Default user")
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

	repo := repository.New(conn)
	secret := key.Secret()
	repo.UpdatePwUsertotpByName(ctx, repository.UpdatePwUsertotpByNameParams{
		Username:   username,
		TotpSecret: []byte(secret),
	})
	// util.JSONResponse(w, secret, http.StatusOK)
	fmt.Println(key.Secret())
}

func CheckSessionToken(w http.ResponseWriter, r *http.Request) (bool, int, Session, string) {
	c, err := r.Cookie("session_token")
	if err != nil {
		if err == http.ErrNoCookie {
			w.WriteHeader(http.StatusUnauthorized)
			return false, http.StatusUnauthorized, Session{}, ""
		}

		w.WriteHeader(http.StatusBadRequest)
		return false, http.StatusBadRequest, Session{}, ""
	}
	sessionToken := c.Value
	repo := repository.New(conn)
	sessions, _ := repo.GetPwUserSessionByUuid(ctx, sessionToken)
	if len(sessions) == 0 {
		w.WriteHeader(http.StatusUnauthorized)
		return false, http.StatusUnauthorized, Session{}, ""
	} else if len(sessions) != 1 {
		w.WriteHeader(http.StatusInternalServerError)
		return false, http.StatusInternalServerError, Session{}, ""
	}
	session := Session{
		username: sessions[0].Username,
		expiry:   sessions[0].ExpiresAt,
	}

	if session.isExpired() {
		l.Println("Session is expired")
		repo.DeletePwUserSessionByUuid(ctx, sessions[0].Uuid)
		w.WriteHeader(http.StatusUnauthorized)
		return false, http.StatusUnauthorized, Session{}, ""
	}

	return true, http.StatusOK, session, session.username
}
