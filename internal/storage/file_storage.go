package storage

import (
	"bufio"
	"context"
	"encoding/json"
	"github.com/Taboon/urlshortner/internal/entity"
	"go.uber.org/zap"
	"io"
	"os"
	"path/filepath"
)

type FileStorage struct {
	fileName string
	Log      *zap.Logger
}

func (f *FileStorage) GetURLsByUser(ctx context.Context, id int) (UserURLs, error) {
	//TODO implement me
	panic("implement me")
}

func (f *FileStorage) GetNewUser(ctx context.Context) (int, error) {
	//TODO implement me
	panic("implement me")
}

func (f *FileStorage) WriteBatchURL(ctx context.Context, b *ReqBatchURLs) (*ReqBatchURLs, error) {
	urlData := URLData{}
	for _, v := range *b {
		urlData.ID = v.ID
		urlData.URL = v.URL
		err := f.AddURL(ctx, urlData)
		if err != nil {
			return nil, err
		}
	}
	return nil, nil
}

func (f *FileStorage) CheckBatchURL(ctx context.Context, urls *ReqBatchURLs) (*ReqBatchURLs, error) {
	for i, v := range *urls {
		_, ok, err := f.CheckURL(ctx, v.URL)
		if err != nil {
			return nil, err
		}
		if ok {
			(*urls)[i].Err = entity.ErrURLExist
		}
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

func (f *FileStorage) Ping(ctx context.Context) error {
	return nil
}

func (f *FileStorage) AddURL(_ context.Context, data URLData) error {
	file, err := os.OpenFile(f.fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer func() {
		err := file.Close()
		if err != nil {
			f.Log.Error("Ошибка закрытия файла", zap.Error(err))
		}
	}()
	encoder := json.NewEncoder(file)
	return encoder.Encode(data)
}

func (f *FileStorage) CheckID(_ context.Context, id string) (URLData, bool, error) {
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
func (f *FileStorage) CheckURL(_ context.Context, url string) (URLData, bool, error) {
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

func (f *FileStorage) RemoveURL(_ context.Context, _ URLData) error { //
	return nil
}
