package domain

import "time"

type User struct {
	UserID    string
	Username  string
	Email     string
	Password  string
	Role      string
	Active    bool
	CreatedAt time.Time
	UpdatedAt time.Time
}
