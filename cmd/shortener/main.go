package main

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	"log"
	"os"

	"github.com/Taboon/urlshortner/internal/config"
	"github.com/Taboon/urlshortner/internal/domain/usecase"
	"github.com/Taboon/urlshortner/internal/logger"
	"github.com/Taboon/urlshortner/internal/server"
	"github.com/Taboon/urlshortner/internal/storage"
	pgx "github.com/jackc/pgx/v5"
)

const baseFilePath = "/tmp/short-url-db.json"

func main() {
	// инициализируем конфиг
	configBuilder := config.NewConfigBuilder()
	configBuilder.SetLocalAddress("127.0.0.1", 8080)
	configBuilder.SetBaseURL("127.0.0.1", 8080)
	configBuilder.SetFileBase(baseFilePath)
	configBuilder.SetLogger("Info")
	configBuilder.ParseEnv()
	configBuilder.ParseFlag()
	conf := configBuilder.Build()

	// инициализируем логгер
	l, err := logger.Initialize(*conf)
	if err != nil {
		log.Fatal(err)
	}

	// инициализируем хранилище
	var stor storage.Repository

	switch {
	case conf.DataBase != "":
		//urlExample := "postgres://postgres:1101@192.168.31.40:5432/urlshortnerdb"
		db, err := pgx.Connect(context.Background(), conf.DataBase)

		if err != nil {
			fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
			os.Exit(1)
		}
		defer db.Close(context.Background())

		stor = storage.NewPostgreBase(db, l)
		err = stor.Ping()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Can't connect to database: %v\n", err)
			os.Exit(1)
		}

		err = storage.Migrations(conf.DataBase)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Can't created table: %v\n", err)
			os.Exit(1)
		}

		l.Info("Использем Postge")
	default:
		stor = storage.NewMemoryStorage(l)
		l.Info("Использем память приложения для хранения")
	}

	//инициализируем URL процессор
	urlProcessor := usecase.URLProcessor{
		Repo: stor,
		Log:  l,
	}

	//инициализируем бекап и загружаем из него данные
	if conf.FileBase.File != "" {
		l.Info("Используем бекап файл", zap.String("file", conf.FileBase.File))
		backuper := storage.NewFileStorage(conf.FileBase.File, l)
		err := backuper.Get(&stor)
		if err != nil {
			log.Fatal(err)
		}
		urlProcessor.Backup = backuper
	}

	//инициализируем сервер
	srv := server.Server{
		LocalAddress: conf.LocalAddress.String(),
		BaseURL:      conf.BaseURL.String(),
		P:            urlProcessor,
		Log: &logger.Logger{
			Logger: l,
		},
	}

	l.Info("Running server", zap.String("address", conf.LocalAddress.String()), zap.String("loglevel", conf.LogLevel))

	if err := srv.Run(conf.LocalAddress); err != nil {
		log.Fatal(err)
	}
}
