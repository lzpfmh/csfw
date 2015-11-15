package dbr

// Builder builds sql in one dialect like MySQL
// e.g. XxxBuilder, Conditioner
type Builder interface {
	Build(Dialect, Buffer) error
}

type BuildFunc func(Dialect, Buffer) error

func (b BuildFunc) Build(d Dialect, buf Buffer) error {
	return b(d, buf)
}
