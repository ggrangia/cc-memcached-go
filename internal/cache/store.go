package cache

type Store interface {
	Get(string) (Data, bool, error)
	Save(Data) error
	Delete(string) error
}
