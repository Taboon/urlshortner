package storage

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"go.uber.org/zap"
	"io"
	"os"
	"path/filepath"
)

type FileStorage struct {
	fileName string
	Log      *zap.Logger
}

var urlData = URLData{}

func (f *FileStorage) AddBatchURL(ctx context.Context, urls map[string]ReqBatchJSON) error {
	for id, v := range urls {
		urlData.ID = id
		urlData.URL = v.URL
		err := f.AddURL(ctx, urlData)
		if err != nil {
			return err
		}
	}
	return nil
}

func (f *FileStorage) CheckBatchURL(ctx context.Context, urls *[]ReqBatchJSON) (*[]ReqBatchJSON, error) {
	for i, v := range *urls {
		_, ok, err := f.CheckURL(ctx, v.URL)
		if err != nil {
			return nil, err
		}
		(*urls)[i].Exist = ok
	}
	return urls, nil
}

var _ Repository = (*FileStorage)(nil)

func NewFileStorage(fileName string, logger *zap.Logger) *FileStorage {
	err := os.MkdirAll(filepath.Dir(fileName), 0774)
	if err != nil {
		logger.Error("Ошибка создания файла")
	}
	logger.Debug("Создали дирректорию", zap.String("dir", filepath.Dir(fileName)))
	return &FileStorage{
		fileName: fileName,
		Log:      logger,
	}
}

func (f *FileStorage) Ping() error {
	return nil
}

func (f *FileStorage) AddURL(ctx context.Context, data URLData) error {
	file, err := os.OpenFile(f.fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	defer func() {
		err := file.Close()
		if err != nil {
			f.Log.Error("Ошибка закрытия файла", zap.Error(err))
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
	return nil
}

func (f *FileStorage) CheckID(ctx context.Context, id string) (URLData, bool, error) {
	file, err := os.OpenFile(f.fileName, os.O_RDONLY|os.O_CREATE|os.O_APPEND, 0774)
	defer func() {
		err := file.Close()
		if err != nil {
			f.Log.Error("Ошибка закрытия файла", zap.Error(err))
		}
	}()

	if err != nil {
		return URLData{}, false, err
	}

	scanner := bufio.NewScanner(file)

	var url *URLData

	for scanner.Scan() {
		err := json.Unmarshal(scanner.Bytes(), &url)
		if err != nil {
			return URLData{}, false, err
		}
		if url.ID == id {
			return *url, true, nil
		}
	}

	return URLData{}, false, nil
}

// Возвращает ok = false если URL нет в базе
func (f *FileStorage) CheckURL(ctx context.Context, url string) (URLData, bool, error) {
	file, err := os.OpenFile(f.fileName, os.O_RDONLY|os.O_CREATE|os.O_APPEND, 0774)
	defer func() {
		err := file.Close()
		if err != nil {
			f.Log.Error("Ошибка закрытия файла", zap.Error(err))
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

func (f *FileStorage) Get(repository *Repository) error {
	file, err := os.OpenFile(f.fileName, os.O_RDONLY|os.O_CREATE, 0774)
	defer func() {
		err := file.Close()
		if err != nil {
			f.Log.Error("Ошибка закрытия файла", zap.Error(err))
		}
	}()
	if err != nil {
		return err
	}

	body, err := io.ReadAll(file)
	if err != nil {
		return err
	}

	if json.Valid(body) {
		err := json.Unmarshal(body, repository)
		if err != nil {
			return err
		}
	}

	return nil
}

func (f *FileStorage) RemoveURL(ctx context.Context, data URLData) error {
	file, err := os.OpenFile(f.fileName, os.O_RDWR, 0774)
	defer func() {
		err := file.Close()
		if err != nil {
			f.Log.Error("Ошибка закрытия файла", zap.Error(err))
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
