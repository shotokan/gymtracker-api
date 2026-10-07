package domain

// GoogleUser represents the user information extracted from a Google ID token.
type GoogleUser struct {
	GoogleID string // sub claim from Google token
	Email    string
	Name     string
	Picture  string // profile picture URL
}
