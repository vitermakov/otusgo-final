package dto

import "errors"

var (
	ErrRequestEmpty = errors.New("empty query")
	ErrBadIP        = errors.New("ip address is not well-formed")
)
