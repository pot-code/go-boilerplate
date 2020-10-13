package domain

import "errors"

// ErrNoSuchUser failed to validate the credential
var ErrNoSuchUser = errors.New("No such user or password is incorrect")

// ErrDuplicatedUser unique key constraint violation
var ErrDuplicatedUser = errors.New("Username or email is already registered")
