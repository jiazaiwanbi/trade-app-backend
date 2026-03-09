package user

import (
	"errors"
	"strings"
	"time"
)

var (
	ErrInvalidEmail    = errors.New("invalid email")
	ErrInvalidNickname = errors.New("invalid nickname")
	ErrUserNotFound    = errors.New("user not found")
)

type User struct {
	ID           int64
	Email        string
	PasswordHash string
	Nickname     string
	Bio          string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func New(email string, passwordHash string, nickname string) (User, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	nickname = strings.TrimSpace(nickname)

	if !strings.Contains(email, "@") {
		return User{}, ErrInvalidEmail
	}
	if nickname == "" {
		return User{}, ErrInvalidNickname
	}

	return User{
		Email:        email,
		PasswordHash: passwordHash,
		Nickname:     nickname,
		Bio:          "",
	}, nil
}

func (u *User) UpdateProfile(nickname string, bio string) error {
	nickname = strings.TrimSpace(nickname)
	if nickname == "" {
		return ErrInvalidNickname
	}

	u.Nickname = nickname
	u.Bio = strings.TrimSpace(bio)
	return nil
}
