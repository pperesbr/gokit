package dsn

type DSN interface {
	Build() (string, error)
}
