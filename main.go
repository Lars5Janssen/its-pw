package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/jackc/pgx/v5"

	"github.com/Lars5Janssen/its-pw/files"
	"github.com/Lars5Janssen/its-pw/httpPages"
	"github.com/Lars5Janssen/its-pw/login"
)

var l log.Logger

func main() {
	// Logging
	l = *log.Default()

	// Connect to DB
	ctx := context.Background()
	conn, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	fmt.Println(os.Getenv("DATABASE_URL"))
	if err != nil {
		os.Exit(1)
	}
	defer conn.Close(ctx)

	// HTTP Server
	http.HandleFunc("/app/", pages.LoginPage)
	http.HandleFunc("/app/login", pages.LoginPage)
	http.HandleFunc("/app/welcome", pages.WelcomePage)
	http.HandleFunc("/app/signup", pages.SignUp)

	http.HandleFunc("/app/beginRegistration", pages.BeginRegistration)
	http.HandleFunc("/app/endRegistration", pages.EndRegistration)

	if _, err := os.Stat("creds.yaml"); err != nil {
		files.WriteYAML("cred.yaml", make(map[string]string))
	}
	pages.InitPasskeys(l, ctx, conn)
	login.AddDefaultUser()

	fmt.Println("Server started")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
