package main

import (
	"github.com/Taboon/urlshortner/internal/config"
	"github.com/Taboon/urlshortner/internal/domain/usecase"
	"github.com/Taboon/urlshortner/internal/logger"
	"github.com/Taboon/urlshortner/internal/server"
	"github.com/Taboon/urlshortner/internal/storage"
	"go.uber.org/zap"
	"log"
)

func main() {
	//инициализируем конфиг
	conf := *config.BuildConfig()

	//инициализируем логгер
	if err := logger.Initialize(conf.LogLevel); err != nil {
		log.Fatal("Can't set logger")
	}

	//инициализируем хранилище
	var stor storage.Repository
	logger.Log.Info("Используем внутреннее хранилище")
	stor = storage.NewMemoryStorage()

	//инициализируем URL процессор
	urlProcessor := usecase.URLProcessor{
		Repo: stor,
	}

	//инициализируем бекап и загружаем из него данные
	if conf.FileBase.File != "" {
		logger.Log.Info("Используем бекап файл", zap.String("file", conf.FileBase.File))
		backuper := storage.NewFileStorage(conf.FileBase.File)
		backuper.Get(&stor)
		urlProcessor.Backup = backuper
	}

	//инициализируем сервер
	srv := server.Server{}
	srv.Conf = &conf
	srv.P = urlProcessor

	logger.Log.Info("Running server", zap.String("address", conf.LocalAddress.String()), zap.String("loglevel", conf.LogLevel))

	if err := srv.Run(); err != nil {
		log.Fatal(err)
	}

}
