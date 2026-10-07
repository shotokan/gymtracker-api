package domain

import "errors"

var (
	// ErrNotFound se retorna cuando no se encuentra un usuario con el ID proporcionado.
	ErrNotFound = errors.New("user not found")
	// ErrAlreadyExists se retorna cuando un email ya está registrado.
	ErrAlreadyExists = errors.New("email already registered")
	// ErrInactive se retorna cuando se intenta acceder a un usuario inactivo.
	ErrInactive = errors.New("user is inactive")
)
