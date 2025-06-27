package pages

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/google/uuid"

	"github.com/Lars5Janssen/its-pw/internal/repository"
)

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
