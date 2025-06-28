package pages

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/google/uuid"

	"github.com/Lars5Janssen/its-pw/internal/repository"
	"github.com/Lars5Janssen/its-pw/util"
)

func BeginRegistration(w http.ResponseWriter, r *http.Request) {
	l.Printf("BeginRegistration\n")
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
		util.JSONResponse(w, msg, http.StatusBadRequest)
		return
	}

	t, err := GenSessionID()
	if err != nil {
		l.Printf("ERROR can't gen session id: %s", err.Error())
		panic(err)
	}

	repoUser, err := repo.GetUserByName(ctx, username)
	util.EasyCheck(err, "ERROR in BeginRegistraition while getting user by name:", "err")
	sessionData, err := json.Marshal(session)
	util.EasyCheck(err, "ERROR in BeginRegistraition while marshaling sessionData:", "err")

	repo.CreateSession(ctx, repository.CreateSessionParams{
		UserID:      repoUser.ID,
		SessionID:   t,
		SessionData: sessionData,
	})

	http.SetCookie(w, &http.Cookie{
		Name:  "sidfix",
		Value: t,
	})
	http.SetCookie(w, &http.Cookie{
		Name:  "name",
		Value: username,
	})

	util.JSONResponse(w, options, http.StatusOK)
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
	marshaledSession, err := repo.GetSessionBySessionId(ctx, sidfix.Value)
	util.EasyCheck(err, "ERROR in EndRegistration while getting session by id:", "err")
	puser := GetUser(name.Value)
	var session webauthn.SessionData
	err = json.Unmarshal(marshaledSession.SessionData, &session)
	util.EasyCheck(err, "ERROR in EndRegistration while unmarshaling session data:", "err")

	credential, err := webAuthn.FinishRegistration(puser, session, r)
	if err != nil {
		msg := fmt.Sprintf("can't finish registration: %s", err.Error())
		l.Printf("ERROR: %s", msg)
		http.SetCookie(w, &http.Cookie{
			Name:  "sid",
			Value: "",
		})
		util.JSONResponse(w, msg, http.StatusBadRequest)
		return
	}

	jsonCreds, err := json.Marshal(credential)
	util.EasyCheck(err, "ERROR in EndRegistration while marshaling creds:", "err")
	user, err := repo.GetUserByName(ctx, name.Value)
	util.EasyCheck(err, "ERROR in EndRegistration while getting user by name:", "err")
	err = repo.UpdateUserCredentials(ctx, repository.UpdateUserCredentialsParams{
		ID:          user.ID,
		Credentials: jsonCreds,
	})
	util.EasyCheck(err, "ERROR in EndRegistration while updating user creds:", "err")
	err = repo.DeleteSessionByUserId(ctx, user.ID)
	util.EasyCheck(err, "ERROR in EndRegistration while deting user session by id:", "err")
	http.SetCookie(w, &http.Cookie{
		Name:  "sid",
		Value: "",
	})

	l.Printf("INFO passkey reg finished")
	util.JSONResponse(w, "Reg Succsess", http.StatusOK)
}
