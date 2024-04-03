package main

import (
	"github.com/Taboon/urlshortner/internal/config"
	"github.com/Taboon/urlshortner/internal/logger"
	"github.com/Taboon/urlshortner/internal/server"
	"github.com/Taboon/urlshortner/internal/storage"
	"go.uber.org/zap"
	"log"
)

func main() {

	conf := *config.BuildConfig()

	if err := logger.Initialize(conf.LogLevel); err != nil {
		log.Fatal("Can't set logger")
	}

	var stor storage.Repository

	if conf.FileBase.File != "" {
		logger.Log.Info("Используем файловое хранилище", zap.String("file", conf.FileBase.File))
		stor = storage.NewFileStorage(conf.FileBase.File)
	} else {
		logger.Log.Info("Используем внутреннее хранилище")
		stor = storage.NewMemoryStorage()
	}

	srv := server.Server{}

	srv.Conf = &conf
	srv.Repo = stor

	logger.Log.Info("Running server", zap.String("address", conf.LocalAddress.String()), zap.String("loglevel", conf.LogLevel))

	if err := srv.Run(); err != nil {
		log.Fatal(err)
	}

}
