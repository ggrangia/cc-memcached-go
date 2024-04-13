package cache

import "sync"

type SimpleStore struct {
	data  map[string]Data
	mutex sync.RWMutex
}

func NewSimpleStore(size int) *SimpleStore {
	return &SimpleStore{
		data: make(map[string]Data, size),
	}
}

func (s *SimpleStore) Get(key string) (Data, bool, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	data, exists := s.data[key]

	return data, exists, nil
}

func (s *SimpleStore) Save(key string, data Data) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.data[key] = data

	return nil
}

func (s *SimpleStore) Delete(key string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	delete(s.data, key)

	return nil
}
