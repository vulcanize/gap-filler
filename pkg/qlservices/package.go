package qlservices

import "errors"

var (
	ErrNoArgs  = errors.New("No arguments")
	ErrBadType = errors.New("Bad argument type")
)
