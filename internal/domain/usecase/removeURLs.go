package usecase

import (
	"context"
	"fmt"
	"github.com/Taboon/urlshortner/internal/storage"
	"go.uber.org/zap"
	"net/http"
	"sync"
)

func (u *URLProcessor) RemoveURLs(requestBody []string, r *http.Request) {
	doneCh := make(chan struct{})

	inputCh := u.generator(doneCh, requestBody)
	channels := u.fanOut(r.Context(), doneCh, inputCh)
	addResultCh := u.fanIn(doneCh, channels...)
	u.remover(r.Context(), doneCh, addResultCh)
}

func (u *URLProcessor) generator(doneCh chan struct{}, input []string) chan string {
	genChan := make(chan string)
	u.Log.Debug("Открыли канал genChan")
	go func() {
		defer func() {
			u.Log.Debug("Закрыли канал genChan")
			close(genChan)
		}()
		fmt.Println(input)

		for _, data := range input {
			select {
			case <-doneCh:
				u.Log.Debug("Получили DONE")
				return
			case genChan <- data:
				u.Log.Debug("Отправили в канал genChan", zap.String("data", data))
			}
		}
	}()

	return genChan
}

func (u *URLProcessor) remover(ctx context.Context, doneCh chan struct{}, idToRemove chan storage.URLData) {
	u.Log.Debug("remover начал работу")
	go func() {
		defer func() {
			close(doneCh)
			u.Log.Debug("Закрыли канал DONE")
		}()

		var batch = make([]storage.URLData, 0)

		for data := range idToRemove {
			batch = append(batch, data)
			u.Log.Debug("Добавили в batch", zap.String("id", data.ID))
		}

		u.Log.Debug("Подготовили batch", zap.Any("batch", batch))
		err := u.Repo.RemoveURL(ctx, batch)
		if err != nil {
			u.Log.Error("Ошибка удаления URL", zap.Error(err))
		}
	}()
}

func (u *URLProcessor) fanOut(ctx context.Context, doneCh chan struct{}, inputCh chan string) []chan storage.URLData {
	numWorkers := 5
	channels := make([]chan storage.URLData, numWorkers)
	for i := 0; i < numWorkers; i++ {
		addResultCh := u.checkID(ctx, doneCh, inputCh)
		channels[i] = addResultCh
	}
	return channels
}

func (u *URLProcessor) checkID(ctx context.Context, doneCh chan struct{}, in chan string) chan storage.URLData {
	checkIDout := make(chan storage.URLData)
	go func() {
		defer func() {
			u.Log.Debug("Закрыли канал checkIDout")
			close(checkIDout)
		}()

		for data := range in {
			url, ok, err := u.Repo.CheckID(ctx, data)
			if err != nil {
				u.Log.Error("ошибка при проверке ID", zap.Error(err))
			}
			u.Log.Debug("Получили инфу по id", zap.String("id", data), zap.Bool("ok", ok), zap.Any("url", url))
			if ok {
				select {
				case <-doneCh:
					return
				case checkIDout <- url:
					u.Log.Debug("Отправили инфу в канал checkIDout")
				}
			}
		}
	}()
	return checkIDout
}

func (u *URLProcessor) fanIn(doneCh chan struct{}, resultChs ...chan storage.URLData) chan storage.URLData {
	finalCh := make(chan storage.URLData)

	var wg sync.WaitGroup

	for _, ch := range resultChs {
		chClosure := ch

		wg.Add(1)
		go func() {
			defer wg.Done()

			for data := range chClosure {
				select {
				case <-doneCh:
					return
				case finalCh <- data:
					u.Log.Debug("Отправляем в канал fanIn", zap.String("id", data.ID))
				}
			}
		}()
	}
	go func() {
		wg.Wait()
		close(finalCh)
		u.Log.Debug("Закрыли канал finalCh")
	}()

	return finalCh
}
