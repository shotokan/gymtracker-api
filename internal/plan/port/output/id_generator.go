package output

// IDGenerator generates unique identifiers.
type IDGenerator interface {
	New() string
}
