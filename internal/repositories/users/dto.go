package users

import (
	"time"

	"github.com/lever-dev/padel-backend/internal/entities"
)

type dto struct {
	ID          string
	PhoneNumber string
	FirstName   string
	LastName    string
	CreatedAt   time.Time
	LastLoginAt *time.Time
}

func newDTO(u *entities.User) dto {
	return dto{
		ID:          u.ID,
		PhoneNumber: u.PhoneNumber,
		FirstName:   u.FirstName,
		LastName:    u.LastName,
		CreatedAt:   u.CreatedAt,
		LastLoginAt: u.LastLoginAt,
	}
}

func (d dto) toEntity() entities.User {
	return entities.User{
		ID:          d.ID,
		PhoneNumber: d.PhoneNumber,
		FirstName:   d.FirstName,
		LastName:    d.LastName,
		CreatedAt:   d.CreatedAt,
		LastLoginAt: d.LastLoginAt,
	}
}
