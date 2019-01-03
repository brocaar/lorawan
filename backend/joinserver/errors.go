package joinserver

import "errors"

// Errors
var (
	ErrInvalidMIC     = errors.New("invalid mic")
	ErrDevEUINotFound = errors.New("deveui does not exist")
)
