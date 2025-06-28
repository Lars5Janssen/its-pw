package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/jackc/pgx/v5"

	"github.com/Lars5Janssen/its-pw/httpPages"
	"github.com/Lars5Janssen/its-pw/login"
	"github.com/Lars5Janssen/its-pw/util"
)

var l log.Logger

func main() {
	// Logging
	l = *log.Default()
	l.Println("VERSION LT DEV")
	util.Init(l)

	// Connect to DB
	ctx := context.Background()
	conn, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		l.Println(os.Getenv("DATABASE_URL"))
		l.Println(err.Error())
		os.Exit(1)
	}
	defer conn.Close(ctx)

	// HTTP Server
	// Redirects
	http.HandleFunc("POST /app/login/proceed", pages.LandingPageRedirect)
	http.HandleFunc("POST /app/proceed", pages.LandingPageRedirect)
	http.HandleFunc("POST /app/", pages.LandingPageRedirect)
	http.HandleFunc("GET /app/proceed", pages.LandingPageRedirect)

	http.HandleFunc("GET /app/LocationTest", pages.LocationTest)

	// Actual Pages
	http.HandleFunc("GET /app/login", pages.LandingPage)
	http.HandleFunc("GET /app/welcome", pages.WelcomePage)

	// Password login endpoints
	http.HandleFunc("POST /app/login", pages.Login)
	http.HandleFunc("POST /app/signup", pages.SignUp)

	// Passkey Login endpoints
	http.HandleFunc("POST /app/beginRegistration", pages.BeginRegistration)
	http.HandleFunc("POST /app/endRegistration", pages.EndRegistration)
	http.HandleFunc("POST /app/beginLogin", pages.BeginLogin)
	http.HandleFunc("POST /app/endLogin", pages.EndLogin)

	login.InitLogin(l, ctx, conn)
	pages.InitPasskeys(l, ctx, conn)
	login.AddDefaultUser()

	l.Println("Server started")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
