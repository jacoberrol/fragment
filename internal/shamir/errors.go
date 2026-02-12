package shamir

import "errors"

var (
	ErrInvalidThreshold = errors.New("invalid threshold")
	ErrInvalidNumShares = errors.New("invalid number of shares")
	ErrInvalidSecret    = errors.New("invalid secret")
)
