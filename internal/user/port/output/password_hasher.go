// internal/user/port/output/password_hasher.go
package output

// PasswordHasher define el contrato de hashing de contraseñas.
type PasswordHasher interface {
	Hash(plain string) (string, error)
	Compare(hash, plain string) error
}
