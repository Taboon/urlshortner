package usecase

import (
	"errors"

	"github.com/Taboon/urlshortner/internal/storage"
)

var errUnknownID = errors.New("отсутствует такой ID")

func (s *URLProcessor) Get(id string) (storage.URLData, error) {
	v, ok, err := s.Repo.CheckID(id)
	if err != nil {
		return v, err
	}
	if !ok {
		return v, errUnknownID
	}
	return v, nil
}
