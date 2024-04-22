package usecase

import (
	"context"
	"go.uber.org/zap"
	"math/rand"
	"strings"

	"github.com/Taboon/urlshortner/internal/entity"
	"github.com/Taboon/urlshortner/internal/storage"
)

const (
	httpPrefix  = "http://"
	httpsPrefix = "https://"
)

func (u *URLProcessor) URLValidator(url string) (string, error) {
	u.Log.Debug("Валидируем URL", zap.String("URL", url))

	url = strings.TrimSpace(url)

	if !strings.HasPrefix(url, httpPrefix) && !strings.HasPrefix(url, httpsPrefix) {
		u.Log.Error("Ошибка валидации URL", zap.String("URL", url))
		return "", entity.ErrHasNoPrefix
	}

	if !strings.Contains(url, ".") {
		return "", entity.ErrHasNoDot
	}

	return url, nil
}

func (u *URLProcessor) SaveURL(ctx context.Context, url string) (string, error) {
	u.Log.Debug("Сохраняем URL")

	data, ok, err := u.Repo.CheckURL(ctx, url)
	if err != nil {
		return "", err
	}

	if ok {
		return data.ID, entity.ErrURLExist
	}

	id := u.generateID(ctx)
	urlObj := storage.URLData{URL: url, ID: id}

	if err := u.Repo.AddURL(ctx, urlObj); err != nil {
		return "", err
	}

	if u.Backup != nil {
		if err := u.Backup.AddURL(ctx, urlObj); err != nil {
			u.Log.Error("Ошибка сохранения бекапа")
		}
	}

	return id, nil
}

func (u *URLProcessor) BatchURLSaver(ctx context.Context, b *storage.ReqBatchURLs) (*storage.ReqBatchURLs, error) {
	u.Log.Info("Пытаемся сохранить массив URL")

	if len(*b) < 1 {
		return nil, entity.ErrNoURLToSave
	}

	var err error

	b = u.batchGenerateID(ctx, b)
	b, err = u.Repo.CheckBatchURL(ctx, u.hasDuplicates(u.validate(b)))
	if err != nil {
		return nil, err
	}

	b, err = u.Repo.AddBatchURL(ctx, b)
	if err != nil {
		return nil, err
	}

	if u.Backup != nil {
		_, err = u.Backup.AddBatchURL(ctx, b)
		if err != nil {
			u.Log.Error("Ошибка сохранения бекапа")
		}
	}

	return b, nil
}

func (u *URLProcessor) hasDuplicates(urls *storage.ReqBatchURLs) *storage.ReqBatchURLs {
	urlMap := make(map[string]bool, len(*urls))

	for i, item := range *urls {
		if _, ok := urlMap[item.URL]; ok {
			(*urls)[i].Err = entity.ErrURLExist // Дубликат найден
		}
		urlMap[item.URL] = true
	}
	return urls
}

func (u *URLProcessor) validate(urls *storage.ReqBatchURLs) *storage.ReqBatchURLs {
	u.Log.Debug("Валидируем массив URL")

	for i, s := range *urls {
		if s.Err != nil {
			continue
		}
		url := strings.TrimSpace(s.URL)

		if !strings.HasPrefix(url, httpPrefix) && !strings.HasPrefix(url, httpsPrefix) {
			u.Log.Info("Ошибка валидации URL. Нет префикса.", zap.String("URL", url))
			(*urls)[i].Err = entity.ErrHasNoPrefix
		}

		if !strings.Contains(url, ".") {
			u.Log.Info("Ошибка валидации URL. Нет точки.", zap.String("URL", url))
			(*urls)[i].Err = entity.ErrHasNoDot
		}
	}
	return urls
}

func (u *URLProcessor) batchGenerateID(ctx context.Context, urls *storage.ReqBatchURLs) *storage.ReqBatchURLs {
	for i := range *urls {
		(*urls)[i].ID = u.generateID(ctx)
	}
	return urls
}

func (u *URLProcessor) generateID(ctx context.Context) string {
	u.Log.Debug("Генерируем ID")

	const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	b := make([]byte, 8)

	for {
		for i := range b {
			if rand.Intn(2) == 0 { //nolint:gosec
				b[i] = letterBytes[rand.Intn(26)] //nolint:gosec    // строчные символы
			} else {
				b[i] = letterBytes[rand.Intn(26)+26] //nolint:gosec    // заглавные символы
			}
		}
		_, ok, err := u.Repo.CheckID(ctx, string(b))
		if err != nil {
			u.Log.Error("Ошибка при проверке ID", zap.Error(err))
		}
		if !ok {
			return string(b)
		}
	}
}
