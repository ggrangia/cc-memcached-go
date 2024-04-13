package cache

type SimpleStore struct {
	data map[string]Data
}

func (s *SimpleStore) Get(key string) (Data, bool, error) {
	data, exists := s.data[key]
	return data, exists, nil
}

func (s *SimpleStore) Save(key string, data Data) error {
	s.data[key] = data

	return nil
}

func (s *SimpleStore) Delete(key string) error {
	delete(s.data, key)
	return nil
}
