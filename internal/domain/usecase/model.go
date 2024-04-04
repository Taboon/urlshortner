package usecase

import (
	"github.com/Taboon/urlshortner/internal/storage"
	"go.uber.org/zap"
)

type URLProcessor struct {
	Repo   storage.Repository
	Backup storage.Repository
	Log    *zap.Logger
}
