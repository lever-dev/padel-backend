package entities

import "time"

type User struct {
	ID          string
	PhoneNumber string
	FirstName   string
	LastName    string
	CreatedAt   time.Time
	LastLoginAt *time.Time
}
