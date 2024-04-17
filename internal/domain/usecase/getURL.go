package usecase

import (
	"context"
	"github.com/Taboon/urlshortner/internal/entity"
	"github.com/Taboon/urlshortner/internal/storage"
)

func (u *URLProcessor) Get(ctx context.Context, id string) (storage.URLData, error) {
	v, ok, err := u.Repo.CheckID(ctx, id)
	if err != nil {
		return v, err
	}
	if !ok {
		return v, entity.ErrUnknownID
	}
	return v, nil
}
