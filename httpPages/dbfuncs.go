package pages

import (
	"encoding/json"

	"github.com/go-webauthn/webauthn/webauthn"

	"github.com/Lars5Janssen/its-pw/internal/repository"
	"github.com/Lars5Janssen/its-pw/passkey"
)

func GetUser(username string) passkey.PasskeyUser {
	repo := repository.New(conn)
	user, _ := repo.GetUserByName(ctx, username)

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
