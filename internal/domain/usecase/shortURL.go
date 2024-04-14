package usecase

import (
	"errors"
	"go.uber.org/zap"
	"math/rand"
	"strings"

	"github.com/Taboon/urlshortner/internal/storage"
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
		return "", errors.New("URL должен начинаться с http:// или https://")
	}
	if !strings.Contains(url, ".") {
		return "", errors.New("is not url")
	}

	return url, nil
}

func (u *URLProcessor) URLSaver(url string) (string, error) {
	u.Log.Debug("Сохраняем URL")
	_, ok, err := u.Repo.CheckURL(url)
	if err != nil {
		return "", err
	}
	if ok {
		return "", errors.New("url already exist")
	}
	id := u.generateID()
	urlObj := storage.URLData{URL: url, ID: id}
	err = u.Repo.AddURL(urlObj)
	if err != nil {
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
