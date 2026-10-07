package domain

import "errors"

// ErrUnauthorized se retorna cuando el token es inválido o no se proporcionó.
var ErrUnauthorized = errors.New("unauthorized")
