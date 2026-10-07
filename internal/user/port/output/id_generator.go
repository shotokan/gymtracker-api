// internal/user/port/output/id_generator.go
package output

// IDGenerator genera identificadores únicos para nuevas entidades.
type IDGenerator interface {
	New() string
}
