package model

import "time"

type User struct {
	ID            string
	Username      string
	PasswordHash  string
	PasswordPlain string
	Role          string
	Name          string
	Email         string
	PhotoURL      string
	GoogleID      string
	IsActive      bool
	CreatedAt     time.Time
	UpdatedAt     time.Time
}
