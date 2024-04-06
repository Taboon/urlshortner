package main

import (
	"go.uber.org/zap"
	"log"

	"github.com/Taboon/urlshortner/internal/config"
	"github.com/Taboon/urlshortner/internal/domain/usecase"
	"github.com/Taboon/urlshortner/internal/server"
	"github.com/Taboon/urlshortner/internal/storage"
)

const baseFilePath = "/tmp/short-url-db.json"

func main() {

	//инициализируем конфиг
	configBuilder := config.NewConfigBuilder()
	configBuilder.SetLocalAddress("127.0.0.1", 8080)
	configBuilder.SetBaseURL("127.0.0.1", 8080)
	configBuilder.SetFileBase(baseFilePath)
	configBuilder.SetLogger("Info")
	configBuilder.ParseEnv()
	configBuilder.ParseFlag()
	conf := configBuilder.Build()

	//инициализируем хранилище
	var stor storage.Repository
	conf.Log.Info("Используем внутреннее хранилище")
	stor = storage.NewMemoryStorage(conf.Log.Logger)

	//инициализируем URL процессор
	urlProcessor := usecase.URLProcessor{
		Repo: stor,
		Log:  conf.Log.Logger,
	}

	//инициализируем бекап и загружаем из него данные
	if conf.FileBase.File != "" {
		conf.Log.Info("Используем бекап файл", zap.String("file", conf.FileBase.File))
		backuper := storage.NewFileStorage(conf.FileBase.File, conf.Log.Logger)
		err := backuper.Get(&stor)
		if err != nil {
			log.Fatal(err)
		}
		urlProcessor.Backup = backuper
	}

	//инициализируем сервер
	srv := server.Server{
		Conf: conf,
		P:    urlProcessor,
		Log:  conf.Log,
	}

	conf.Log.Info("Running server", zap.String("address", conf.LocalAddress.String()), zap.String("loglevel", conf.Log.LogLevel))

	if err := srv.Run(); err != nil {
		log.Fatal(err)
	}

}
