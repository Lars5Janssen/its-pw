// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.29.0

package repository

import (
	"time"
)

type Pwsession struct {
	Username  string
	Uuid      string
	ExpiresAt time.Time
}

type Pwuser struct {
	Username   string
	Pw         []byte
	TotpSecret []byte
}

type Session struct {
	UserID      []byte
	SessionID   string
	SessionData []byte
}

type User struct {
	ID          []byte
	DisplayName string
	Name        string
	Credentials []byte
}
