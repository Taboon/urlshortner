package main

import (
	"context"
	"github.com/Taboon/urlshortner/internal/server/auth"
	"log"

	"github.com/Taboon/urlshortner/internal/config"
	"github.com/Taboon/urlshortner/internal/domain/usecase"
	"github.com/Taboon/urlshortner/internal/logger"
	"github.com/Taboon/urlshortner/internal/server"
	"github.com/Taboon/urlshortner/internal/storage"
	pgx "github.com/jackc/pgx/v5"
	"go.uber.org/zap"
)

func main() { //nolint:funlen
	conf := config.SetConfig()

	// инициализируем логгер
	l, err := logger.Initialize(*conf)
	if err != nil {
		log.Fatal(err)
	}

	// инициализируем хранилище
	var stor storage.Repository

	switch {
	case conf.DataBase != "":
		db, s := storage.SetPostgres(conf, l)
		stor = s
		defer func(db *pgx.Conn, ctx context.Context) {
			err := db.Close(ctx)
			if err != nil {
				panic(err)
			}
		}(db, context.Background())
	default:
		stor = storage.NewMemoryStorage(l)
		l.Info("Используем память приложения для хранения")
	}

	// инициализируем URL процессор
	urlProcessor := usecase.URLProcessor{
		Repo:            stor,
		Log:             l,
		Authentificator: auth.NewAuthentificator(l, stor),
	}

	// инициализируем бекап и загружаем из него данные
	if conf.FileBase.File != "" {
		l.Info("Используем бекап файл", zap.String("file", conf.FileBase.File))
		backuper := storage.NewFileStorage(conf.FileBase.File, l)
		err := backuper.Get(&stor)
		if err != nil {
			panic(err)
		}
		urlProcessor.Backup = backuper
	}

	// инициализируем сервер
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
		panic(err)
	}
}
