package usecase

import (
	"go.uber.org/zap"

	"github.com/Taboon/urlshortner/internal/storage"
)

type URLProcessor struct {
	Repo   storage.Repository
	Backup storage.Repository
	Log    *zap.Logger
}
