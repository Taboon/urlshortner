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

var _ Repositories = (*TempStorage)(nil)
var sm = SafeMap{}

func (s TempStorage) AddURL(data URLData) error {
	err := sm.Write(data)
	if err != nil {
		return err
	}
	return nil
}

func mapChecker() {
	if sm.mapStor == nil {
		sm.mapStor = make(map[string]URLData)
	}
	if sm.reverseMapStor == nil {
		sm.reverseMapStor = make(map[URLData]string)
	}
}

func (s TempStorage) CheckID(id string) (URLData, bool) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// Проверяем, что map был инициализирован
	mapChecker()

	v, ok := sm.mapStor[id]
	if ok {
		return v, true
	}
	return URLData{}, false
}

func (s TempStorage) CheckURL(url string) (URLData, bool) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// Проверяем, что map был инициализирован
	mapChecker()

	for _, v := range sm.mapStor {
		if v.URL == url {
			return v, true
		}
	}
	return URLData{}, false
}

func (s TempStorage) RemoveURL(data URLData) error {
	return sm.Remove(data)
}

func (sm *SafeMap) Write(url URLData) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// Проверяем, что map был инициализирован
	mapChecker()

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
	mapChecker()

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
