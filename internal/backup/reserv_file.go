package backup

import (
	"encoding/json"
	"github.com/Taboon/urlshortner/internal/logger"
	"github.com/Taboon/urlshortner/internal/storage"
	"go.uber.org/zap"
	"io"
	"os"
)

type ReservFile struct {
	FileName string
}

func (r *ReservFile) Get(repository *storage.Repository) error {
	file, err := os.OpenFile(r.FileName, os.O_RDONLY|os.O_CREATE, 0774)
	defer func() {
		err := file.Close()
		if err != nil {
			logger.Log.Error("Ошибка закрытия файла", zap.Error(err))
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

func (r *ReservFile) Save(repository *storage.Repository) error {
	file, err := os.Create(r.FileName)
	defer func() {
		err := file.Close()
		if err != nil {
			logger.Log.Error("Ошибка закрытия файла", zap.Error(err))
		}
	}()
	if err != nil {
		return err
	}

	body, err := json.Marshal(repository)
	if err != nil {
		return err
	}

	_, err = file.Write(body)
	if err != nil {
		return err
	}

	return nil
}
