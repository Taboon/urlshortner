package usecase

import (
	"errors"
	"fmt"
	"github.com/Taboon/urlshortner/internal/logger"
	"github.com/Taboon/urlshortner/internal/storage"
	"go.uber.org/zap"
	"math/rand"
	"strings"
)

const (
	httpPrefix  = "http://"
	httpsPrefix = "https://"
)

func (s *URLProcessor) UrlValidator(url string) (string, error) {
	logger.Log.Debug("Валидируем URL")
	url = strings.TrimSpace(url)
	if !strings.HasPrefix(url, httpPrefix) && !strings.HasPrefix(url, httpsPrefix) {
		fmt.Println("Это не URL - не указан http:// или https://")
		return "", errors.New("URL должен начинаться с http:// или https://")
	}
	if !strings.Contains(url, ".") {
		return "", errors.New("is not url")
	}

	return url, nil
}

func (s *URLProcessor) URLSaver(url string) (string, error) {
	logger.Log.Debug("Сохраняем URL")
	_, ok, err := s.Repo.CheckURL(url)
	if err != nil {
		return "", err
	}
	if ok {
		return "", errors.New("url already exist")
	}
	id := s.generateID()
	urlObj := storage.URLData{URL: url, ID: id}
	err = s.Repo.AddURL(urlObj)
	if err != nil {
		return "", err
	}
	return id, nil
}

func (s *URLProcessor) generateID() string {
	logger.Log.Debug("Генерируем ID")
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
		_, ok, err := s.Repo.CheckID(string(b))
		if err != nil {
			logger.Log.Error("Ошибка при проверке ID", zap.Error(err))
		}
		if !ok {
			return string(b)
		}
	}
}