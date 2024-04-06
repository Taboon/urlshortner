package usecase

import (
	"github.com/Taboon/urlshortner/internal/entity"
	"github.com/Taboon/urlshortner/internal/storage"
)

func (u *URLProcessor) Get(id string) (storage.URLData, error) {
	v, ok, err := u.Repo.CheckID(id)
	if err != nil {
		return v, err
	}
	if !ok {
		return v, entity.ErrUnknownID
	}
	return v, nil
}
