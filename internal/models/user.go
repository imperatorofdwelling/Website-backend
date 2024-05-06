package models

import "github.com/google/uuid"

type User struct {
	Id      uuid.UUID `yaml:"id"`
	Balance float64   `json:"balance"`
}

func NewUser() *User {
	return &User{
		Id:      uuid.New(),
		Balance: 0,
	}
}
