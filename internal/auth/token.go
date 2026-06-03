package auth

import "github.com/google/uuid"

func NewToken() string {
	return uuid.New().String()
}
