package usecase

import (
	"errors"
	"github.com/Taboon/urlshortner/internal/storage"
	"go.uber.org/zap"
	"math/rand"
	"strings"

	"github.com/Taboon/urlshortner/internal/entity"
)

const (
	httpPrefix  = "http://"
	httpsPrefix = "https://"
)

func (u *URLProcessor) URLValidator(url string) (string, error) {
	u.Log.Debug("Валидируем URL", zap.String("url", url))

	url = strings.TrimSpace(url)

	if !strings.HasPrefix(url, httpPrefix) && !strings.HasPrefix(url, httpsPrefix) {
		u.Log.Error("Ошибка валидации URL", zap.String("url", url))
		return "", entity.ErrURLInvalid
	}

	if !strings.Contains(url, ".") {
		return "", entity.ErrIsNoURL
	}

	return url, nil
}

func (u *URLProcessor) BatchURLValidator(urls *[]storage.ReqBatchJSON) *[]storage.ReqBatchJSON {
	u.Log.Debug("Валидируем массив URL")

	for i, s := range *urls {
		(*urls)[i].Valid = true
		url := strings.TrimSpace(s.URL)

		if !strings.HasPrefix(url, httpPrefix) && !strings.HasPrefix(url, httpsPrefix) {
			u.Log.Info("Ошибка валидации URL. Нет префикса.", zap.String("url", url))
			(*urls)[i].Valid = false
		}

		if !strings.Contains(url, ".") {
			u.Log.Info("Ошибка валидации URL. Нет точки.", zap.String("url", url))
			(*urls)[i].Valid = false
		}
	}
	return urls
}

func (u *URLProcessor) URLSaver(url string) (string, error) {
	u.Log.Debug("Сохраняем URL")

	data, ok, err := u.Repo.CheckURL(url)

	if err != nil {
		return "", err
	}

	if ok {
		return data.ID, entity.ErrURLExist
	}

	id := u.generateID()

	urlObj := storage.URLData{URL: url, ID: id}

	err = u.Repo.AddURL(urlObj)

	if err != nil {
		if errors.Is(err, entity.ErrURLExist) {
			return "", err
		}
		return "", err
	}

	if u.Backup != nil {
		err := u.Backup.AddURL(urlObj)
		if err != nil {
			u.Log.Error("Ошибка сохранения бекапа")
		}
	}

	return id, nil
}

func (u *URLProcessor) BatchURLSaver(urls *[]storage.ReqBatchJSON) (map[string]storage.ReqBatchJSON, error) {
	u.Log.Debug("Сохраняем массив URL")
	urlsChecked, err := u.Repo.CheckBatchURL(urls)
	if err != nil {
		return nil, err
	}
	urls = urlsChecked

	var urlsToDB = make(map[string]storage.ReqBatchJSON)
	var urlsWithErr = make(map[string]storage.ReqBatchJSON)

	for _, v := range *urls {
		id := u.generateID()
		if v.Valid && !v.Exist {
			urlsToDB[id] = v
		} else {
			urlsWithErr[id] = v
		}
	}

	u.Log.Info("Пытаемся сохранить массив URL")
	if len(urlsToDB) > 0 {
		err = u.Repo.AddBatchURL(urlsToDB)
		if err != nil {
			return nil, err
		}

		if u.Backup != nil {
			err := u.Backup.AddBatchURL(urlsToDB)
			if err != nil {
				u.Log.Error("Ошибка сохранения бекапа")
			}
		}
	}

	for i, v := range urlsWithErr {
		urlsToDB[i] = v
	}

	return urlsToDB, nil
}

func (u *URLProcessor) generateID() string {
	u.Log.Debug("Генерируем ID")

	const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	b := make([]byte, 8)

	for {
		for i := range b {
			if rand.Intn(2) == 0 {
				b[i] = letterBytes[rand.Intn(26)] // строчные символы
			} else {
				b[i] = letterBytes[rand.Intn(26)+26] // заглавные символы
			}
		}
		_, ok, err := u.Repo.CheckID(string(b))
		if err != nil {
			u.Log.Error("Ошибка при проверке ID", zap.Error(err))
		}
		if !ok {
			return string(b)
		}
	}
}
