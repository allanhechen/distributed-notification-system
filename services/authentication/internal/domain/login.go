package domain

import (
	"errors"

	"github.com/google/uuid"
)


var LoginAttemptFailed = errors.New("Too many invalid logins")
var CredentialIncorrect = errors.New("Incorrect username or password")


type user struct {
	ClientId uuid
	Username string
	HashedPassword string
}