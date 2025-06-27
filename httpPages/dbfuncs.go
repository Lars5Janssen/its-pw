package pages

import (
	"github.com/go-webauthn/webauthn/webauthn"

	"github.com/Lars5Janssen/its-pw/internal/repository"
	"github.com/Lars5Janssen/its-pw/passkey"
)

func CreateUser(username string) {}
func GetUser(username string) passkey.PasskeyUser {
	repo := repository.New(conn)
	user, _ := repo.GetUserByName(ctx, username)

	u := passkey.User{
		ID:          user.ID,
		Name:        user.Name,
		DisplayName: user.DisplayName,
	}
	credentials, _ := repo.GetUserCredsById(ctx, user.ID)
	creds := webauthn.Credential{
		ID:              credentials.ID,
		PublicKey:       credentials.PublicKey,
		AttestationType: credentials.AttestationType,
		Transport:       credentials.Transport,
		Flags:           credentials.Flags,
		Authenticator:   credentials.Authenticator,
		Attestation:     credentials.Attestation,
	}
	u.AddCredential(&creds)

	return nil
}
