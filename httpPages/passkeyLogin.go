package pages

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Lars5Janssen/its-pw/internal/repository"
	"github.com/go-webauthn/webauthn/webauthn"
)

func BeginLogin(w http.ResponseWriter, r *http.Request) {
	l.Println("Begin Login")
	username, err := getUsername(r)
	if err != nil {
		l.Printf("ERROR can't get user name: %s", err.Error())
		panic(err)
	}

	user := GetUser(username)

	options, session, err := webAuthn.BeginLogin(user)
	if err != nil {
		msg := fmt.Sprintf("can't begin login: %s", err.Error())
		l.Printf("ERROR %s", msg)
		JSONResponse(w, msg, http.StatusBadRequest)

		return
	}
	t, err := GenSessionID()
	if err != nil {
		l.Printf("ERROR can't generate session id: %s", err.Error())
		panic(err)
	}

	repo := repository.New(conn)
	repoUser, _ := repo.GetUserByName(ctx, username)
	sessionData, _ := json.Marshal(session)
	repo.CreateSession(ctx, repository.CreateSessionParams{
		UserID: repoUser.ID,
		SessionID: t,
		SessionData: sessionData,
	})

	http.SetCookie(w, &http.Cookie{
		Name:     "sid",
		Value:    t,
		Path:     "api/passkey/loginStart",
		MaxAge:   3600,
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode, // TODO: SameSiteStrictMode maybe?
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

func EndLogin(w http.ResponseWriter, r *http.Request) {
	l.Println("End Login")
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
	marshaledSession, _ := repo.GetSessionBySessionId(ctx, sidfix.Value)
	puser := GetUser(name.Value)
	var session webauthn.SessionData
	json.Unmarshal(marshaledSession.SessionData, &session)

	credential, err := webAuthn.FinishLogin(puser, session, r)
	if err != nil {
		l.Printf("can't finish login: %s", err.Error())
		panic(err)
	}

	if credential.Authenticator.CloneWarning {
		l.Println("WARNING can't finish login due to Clone Warining")
	}




	jsonCreds, _ := json.Marshal(credential)
	user, _ := repo.GetUserByName(ctx, name.Value)
	repo.UpdateUserCredentials(ctx, repository.UpdateUserCredentialsParams{
		ID:          user.ID,
		Credentials: jsonCreds,
	})
	repo.DeleteSessionByUserId(ctx, user.ID)
	http.SetCookie(w, &http.Cookie{
		Name:  "sid",
		Value: "",
	})

	setLoginSessionToken(w, user.Name)

	fmt.Printf("User %s logged in\n", user.Name)
	JSONResponse(w, "LOGIN Success", http.StatusOK)
	// http.Redirect(w, r, "/app/welcome", http.StatusSeeOther)
	return
}
