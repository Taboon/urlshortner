package storage

import (
	"errors"
	"sync"
)

type URLData struct {
	URL string
	ID  string
}

type SafeMap struct {
	mu             sync.Mutex
	mapStor        map[string]URLData
	reverseMapStor map[URLData]string
}

type TempStorage struct {
	sm *SafeMap
}

//var _ Repository = (*TempStorage)(nil)

func NewTempStorage() Repository {
	var data = SafeMap{}
	return TempStorage{
		sm: &data,
	}
}

func mapChecker(sm *SafeMap) {
	if sm.mapStor == nil {
		sm.mapStor = make(map[string]URLData)
	}
	if sm.reverseMapStor == nil {
		sm.reverseMapStor = make(map[URLData]string)
	}
}

func (s TempStorage) AddURL(data URLData) error {
	err := s.sm.Write(data)
	if err != nil {
		return err
	}
	return nil
}

func (s TempStorage) CheckID(id string) (URLData, bool) {
	s.sm.mu.Lock()
	defer s.sm.mu.Unlock()

	// Проверяем, что map был инициализирован
	mapChecker(s.sm)

	v, ok := s.sm.mapStor[id]
	if ok {
		return v, true
	}
	return URLData{}, false
}

func (s TempStorage) CheckURL(url string) (URLData, bool) {
	s.sm.mu.Lock()
	defer s.sm.mu.Unlock()

	// Проверяем, что map был инициализирован
	mapChecker(s.sm)

	for _, v := range s.sm.mapStor {
		if v.URL == url {
			return v, true
		}
	}
	return URLData{}, false
}

func (s TempStorage) RemoveURL(data URLData) error {
	return s.sm.Remove(data)
}

func (sm *SafeMap) Write(url URLData) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// Проверяем, что map был инициализирован
	mapChecker(sm)

	// Пишем данные в map
	_, ok := sm.mapStor[url.ID]
	if ok {
		err := errors.New("id exist")
		return err
	}

	_, ok = sm.reverseMapStor[url]
	if ok {
		err := errors.New("url exist")
		return err
	}

	sm.mapStor[url.ID] = url
	sm.reverseMapStor[url] = url.ID
	return nil

}

func (sm *SafeMap) Remove(url URLData) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// Проверяем, что map был инициализирован
	mapChecker(sm)

	// Удаляем данные из map
	_, ok := sm.mapStor[url.ID]
	if ok {
		_, ok := sm.reverseMapStor[url]
		if ok {
			delete(sm.mapStor, url.ID)
			delete(sm.reverseMapStor, url)
			return nil
		}
	}
	err := errors.New("removing error")
	return err
}
