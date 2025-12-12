package repository

import "errors"

// ErrNoRows signifies that no rows were found associated with the query
var ErrNoRows = errors.New("repository: not found")

// ErrAlreadyExists signifies that the desired insertion violates
// uniqueness constraints within the database
var ErrAlreadyExists = errors.New("repository: entity already exists")
