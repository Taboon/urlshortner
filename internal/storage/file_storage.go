package storage

import (
	"bufio"
	"encoding/json"
	"errors"
	"github.com/Taboon/urlshortner/internal/logger"
	"go.uber.org/zap"
	"os"
)

type FileStorage struct {
	fileName string
}

var _ Repository = (*FileStorage)(nil)

func NewFileStorage(fileName string) *FileStorage {
	return &FileStorage{
		fileName: fileName,
	}
}

func (f *FileStorage) AddURL(data URLData) error {
	file, err := os.OpenFile(f.fileName, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	defer func() {
		err := file.Close()
		if err != nil {
			logger.Log.Error("Ошибка закрытия файла", zap.Error(err))
		}
	}()
	if err != nil {
		return err
	}
	encoder := json.NewEncoder(file)
	err = encoder.Encode(data)
	if err != nil {
		return err
	}
	err = file.Close()
	if err != nil {
		return err
	}
	return nil
}

func (f *FileStorage) CheckID(id string) (URLData, bool, error) {
	file, err := os.OpenFile(f.fileName, os.O_RDONLY|os.O_CREATE|os.O_APPEND, 0666)
	defer func() {
		err := file.Close()
		if err != nil {
			logger.Log.Error("Ошибка закрытия файла", zap.Error(err))
		}
	}()

	if err != nil {
		return URLData{}, false, err
	}
	scanner := bufio.NewScanner(file)
	var shortUrls []URLData

	for scanner.Scan() {
		var url URLData
		err := json.Unmarshal(scanner.Bytes(), &url)
		if err != nil {
			return URLData{}, false, err
		}
		shortUrls = append(shortUrls, url)
	}

	for _, v := range shortUrls {
		if v.ID == id {
			return v, true, nil
		}
	}

	return URLData{}, false, nil
}

func (f *FileStorage) CheckURL(url string) (URLData, bool, error) {
	file, err := os.OpenFile(f.fileName, os.O_RDONLY|os.O_CREATE|os.O_APPEND, 0666)
	defer func() {
		err := file.Close()
		if err != nil {
			logger.Log.Error("Ошибка закрытия файла", zap.Error(err))
		}
	}()

	if err != nil {
		return URLData{}, false, err
	}
	scanner := bufio.NewScanner(file)
	var shortUrls []URLData

	for scanner.Scan() {
		var url URLData
		err := json.Unmarshal(scanner.Bytes(), &url)
		if err != nil {
			return URLData{}, false, err
		}
		shortUrls = append(shortUrls, url)
	}

	for _, v := range shortUrls {
		if v.URL == url {
			return v, true, nil
		}
	}

	return URLData{}, false, nil
}

func (f *FileStorage) RemoveURL(data URLData) error {
	file, err := os.OpenFile(f.fileName, os.O_RDWR, 0666)
	defer func() {
		err := file.Close()
		if err != nil {
			logger.Log.Error("Ошибка закрытия файла", zap.Error(err))
		}
	}()

	if err != nil {
		return err
	}
	scanner := bufio.NewScanner(file)
	var shortUrls []URLData

	for scanner.Scan() {
		var url URLData
		err := json.Unmarshal(scanner.Bytes(), &url)
		if err != nil {
			return err
		}
		shortUrls = append(shortUrls, url)
	}

	var newURLs []URLData
	urlForRemove := false

	for _, v := range shortUrls {
		if v == data {
			urlForRemove = true
		} else {
			newURLs = append(newURLs, v)
		}
	}

	if urlForRemove {
		return nil
	}

	body, err := json.Marshal(newURLs)
	if err != nil {
		return err
	}

	_, err = file.Write(body)
	if err != nil {
		return err
	}

	return errors.New("id not found")
}
