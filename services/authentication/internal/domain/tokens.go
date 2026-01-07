package domain

import (
	"errors"
)

var ExpiryRefresh uint = 1800

var ExpiryAccessError = errors.New("Access token has been expired")
var ExpiryRefreshError = errors.New("Refresh token has been expired")



// this can be both refresh and access token type
type Token struct {

	HashedToken string
	Lifespan uint
	TokenType uint
	
}