package pages

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"log"

	"github.com/go-webauthn/webauthn/webauthn"
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
			"http://localhost:8765",
			"https://localhost:8080/app/",
			"localhost:8080",
		},
	}

	if webAuthn, err = webauthn.New(wconfig); err != nil {
		log.Fatalln(err)
	}
}

func GenSessionID() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	return base64.URLEncoding.EncodeToString(b), nil
}
