package storage

import (
	"bufio"
	"context"
	"encoding/json"
	"github.com/Taboon/urlshortner/internal/entity"
	"go.uber.org/zap"
	"os"
	"path/filepath"
)

type FileStorage struct {
	fileName string
	Log      *zap.Logger
}

type URLInFile struct {
	ID     string `json:"id"`
	URL    string `json:"url"`
	UserID int    `json:"user_id"`
}

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

func (f *FileStorage) Set(url URLInFile) error {
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
	return encoder.Encode(url)
}

func (f *FileStorage) Get(repository *InternalStorage) error {
	if repository == nil {
		return entity.ErrRepositoryNotInitialized
	}

	var data = URLInFile{}
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

	scanner := bufio.NewScanner(file)

	// Читаем файл построчно
	for scanner.Scan() {
		line := scanner.Bytes()
		if json.Valid(line) {
			err := json.Unmarshal(line, &data)
			if err != nil {
				return err
			}
		}
		repository.Users[data.UserID] = append(UserURLs{}, URLData{ID: data.ID, URL: data.URL})
	}

	return nil
}

func (f *FileStorage) RemoveURL(_ context.Context, _ URLData) error { //
	return nil
}
