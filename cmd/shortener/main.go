package main

import (
	"go.uber.org/zap"
	"log"

	"github.com/Taboon/urlshortner/internal/config"
	"github.com/Taboon/urlshortner/internal/domain/usecase"
	"github.com/Taboon/urlshortner/internal/server"
	"github.com/Taboon/urlshortner/internal/storage"
)

func main() {
	//инициализируем конфиг
	conf := *config.BuildConfig()

	//инициализируем хранилище
	var stor storage.Repository
	//conf.Log.Info("Используем внутреннее хранилище")
	stor = storage.NewPostgreBase("urlshortnerdb", "postgres", "1101", "192.168.31.40", "5432", conf.Log)

	//инициализируем URL процессор
	urlProcessor := usecase.URLProcessor{
		Repo: stor,
		Log:  conf.Log,
	}

	//инициализируем бекап и загружаем из него данные
	if conf.FileBase.File != "" {
		conf.Log.Info("Используем бекап файл", zap.String("file", conf.FileBase.File))
		backuper := storage.NewFileStorage(conf.FileBase.File, conf.Log)
		err := backuper.Get(&stor)
		if err != nil {
			log.Fatal(err)
		}
		urlProcessor.Backup = backuper
	}

	//инициализируем сервер
	srv := server.Server{}
	srv.Conf = &conf
	srv.P = urlProcessor

	conf.Log.Info("Running server", zap.String("address", conf.LocalAddress.String()), zap.String("loglevel", conf.LogLevel))
	stor.Ping()
	if err := srv.Run(); err != nil {
		log.Fatal(err)
	}

}
