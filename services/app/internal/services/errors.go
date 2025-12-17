package services

import "errors"

// ErrNotFound signifies that no objects were found associated from the
// repository
var ErrNotFound = errors.New("service: not found")

// ErrExpired signifies that the found object from the repository was
// marked as expired
var ErrExpired = errors.New("service: row marked expired")

// ErrFailed signifies that the IdempotentRequest was marked as previously
// failed
var ErrFailed = errors.New("service: idempotent request marked failed")

// ErrConflict signifies that a conflict occurred with another request
var ErrConflict = errors.New("service: conflict with other request")
