package storage

import (
	"errors"
	"go.uber.org/zap"
	"sync"
)

type SafeMap struct {
	mu             sync.Mutex
	mapStor        map[string]string
	reverseMapStor map[string]string
	Log            *zap.Logger
}

var _ Repository = (*SafeMap)(nil)

func NewMemoryStorage(logger *zap.Logger) *SafeMap {
	return &SafeMap{
		mu:             sync.Mutex{},
		mapStor:        make(map[string]string),
		reverseMapStor: make(map[string]string),
		Log:            logger,
	}
}

func (sm *SafeMap) AddURL(data URLData) error {
	sm.Log.Debug("Сохраняем URL")
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// Проверяем, что map был инициализирован
	sm.mapChecker()

	// Пишем данные в map
	_, ok := sm.mapStor[data.ID]
	if ok {
		err := errors.New("id exist")
		return err
	}

	_, ok = sm.reverseMapStor[data.URL]
	if ok {
		err := errors.New("url exist")
		return err
	}

	sm.mapStor[data.ID] = data.URL
	sm.reverseMapStor[data.URL] = data.ID
	return nil
}

func (sm *SafeMap) CheckID(id string) (URLData, bool, error) {
	sm.Log.Debug("Проверяем ID")
	urlData := URLData{}
	val, ok := sm.mapStor[id]
	if ok {
		urlData.ID = id
		urlData.URL = val
		return urlData, true, nil
	}
	return urlData, false, nil
}

func (sm *SafeMap) CheckURL(url string) (URLData, bool, error) {
	sm.Log.Debug("Проверяем URL")
	urlData := URLData{}
	val, ok := sm.reverseMapStor[url]
	if ok {
		urlData.ID = val
		urlData.URL = url
		return urlData, true, nil
	}
	return urlData, false, nil
}

func (sm *SafeMap) RemoveURL(data URLData) error {
	sm.Log.Debug("Удаляем URL")
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// Проверяем, что map был инициализирован
	sm.mapChecker()

	// Удаляем данные из map
	_, ok := sm.mapStor[data.ID]
	if ok {
		_, ok := sm.reverseMapStor[data.URL]
		if ok {
			delete(sm.mapStor, data.ID)
			delete(sm.reverseMapStor, sm.mapStor[data.URL])
			return nil
		}
	}
	err := errors.New("removing error")
	return err
}

func (sm *SafeMap) mapChecker() {
	if sm.mapStor == nil {
		sm.mapStor = make(map[string]string)
	}
	if sm.reverseMapStor == nil {
		sm.reverseMapStor = make(map[string]string)
	}
}
