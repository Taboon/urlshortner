package storage

import (
	"errors"
	"sync"
)

type UrlData struct {
	Url string
	Id  string
}

type SafeMap struct {
	mu             sync.Mutex
	mapStor        map[string]UrlData
	reverseMapStor map[UrlData]string
}

var _ Repositories = (*TempStorage)(nil)
var sm = SafeMap{}

func (s TempStorage) AddUrl(data UrlData) error {
	err := sm.Write(data)
	if err != nil {
		return err
	}
	return nil
}

func (s TempStorage) CheckId(id string) (UrlData, bool) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// Проверяем, что map был инициализирован
	if sm.mapStor == nil {
		sm.mapStor = make(map[string]UrlData)
	}
	if sm.reverseMapStor == nil {
		sm.reverseMapStor = make(map[UrlData]string)
	}

	v, ok := sm.mapStor[id]
	if ok {
		return v, true
	}
	return UrlData{}, false
}

func (s TempStorage) CheckUrl(url string) (UrlData, bool) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// Проверяем, что map был инициализирован
	if sm.mapStor == nil {
		sm.mapStor = make(map[string]UrlData)
	}
	if sm.reverseMapStor == nil {
		sm.reverseMapStor = make(map[UrlData]string)
	}

	for _, v := range sm.mapStor {
		if v.Url == url {
			return v, true
		}
	}
	return UrlData{}, false
}

func (s TempStorage) RemoveUrl(data UrlData) error {
	err := sm.Remove(data)
	if err != nil {
		return err
	}
	return nil
}

func (sm *SafeMap) Write(url UrlData) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// Проверяем, что map был инициализирован
	if sm.mapStor == nil {
		sm.mapStor = make(map[string]UrlData)
	}
	if sm.reverseMapStor == nil {
		sm.reverseMapStor = make(map[UrlData]string)
	}

	// Пишем данные в map
	_, ok := sm.mapStor[url.Id]
	if ok {
		err := errors.New("id exist")
		return err
	} else {
		_, ok := sm.reverseMapStor[url]
		if ok {
			err := errors.New("url exist")
			return err
		} else {
			sm.mapStor[url.Id] = url
			sm.reverseMapStor[url] = url.Id
			return nil
		}
	}
}

func (sm *SafeMap) Remove(url UrlData) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// Проверяем, что map был инициализирован
	if sm.mapStor == nil {
		sm.mapStor = make(map[string]UrlData)
	}
	if sm.reverseMapStor == nil {
		sm.reverseMapStor = make(map[UrlData]string)
	}

	// Удаляем данные из map
	_, ok := sm.mapStor[url.Id]
	if ok {
		_, ok := sm.reverseMapStor[url]
		if ok {
			delete(sm.mapStor, url.Id)
			delete(sm.reverseMapStor, url)
			return nil
		}
	}
	err := errors.New("removing error")
	return err
}
