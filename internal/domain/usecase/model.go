package usecase

import (
	"github.com/Taboon/urlshortner/internal/server/auth"
	"go.uber.org/zap"

	"github.com/Taboon/urlshortner/internal/storage"
)

type URLProcessor struct {
	Repo            storage.Repository
	Authentificator auth.Autentificator
	Log             *zap.Logger
}
