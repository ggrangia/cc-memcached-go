package cache

type Store interface {
	Get(string) (Data, bool, error)
	Save(string, Data) error
	Delete(string) error
}
