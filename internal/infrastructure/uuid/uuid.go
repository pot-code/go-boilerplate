package uuid

type UUIDGenerator interface {
	Generate() (string, error)
}
