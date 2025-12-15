package services

import "errors"

// ErrNotFound signifies that no objects were found associated from the
// repository
var ErrNotFound = errors.New("service: not found")

// ErrConflict signifies that a conflict
var ErrConflict = errors.New("service: conflict with other request")
