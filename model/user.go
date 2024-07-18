package model

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Username   string
	Role       string
	Hash       []byte
	CreatedAt  time.Time
	ModifiedAt time.Time
}
