package usecase

import "github.com/Taboon/urlshortner/internal/storage"

type URLProcessor struct {
	Repo   storage.Repository
	Backup storage.Repository
}
