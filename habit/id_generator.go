package habit

type IDGenerator interface {
	NewID() (string, error)
}
