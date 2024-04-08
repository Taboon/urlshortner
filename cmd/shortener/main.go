package main

import (
	"context"
	"fmt"
	"github.com/Taboon/urlshortner/internal/config"
	"github.com/Taboon/urlshortner/internal/domain/usecase"
	"github.com/Taboon/urlshortner/internal/logger"
	"github.com/Taboon/urlshortner/internal/server"
	"github.com/Taboon/urlshortner/internal/storage"
	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
	"log"
	"os"
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

	//инициализируем логгер
	l, err := logger.Initialize(*conf)
	if err != nil {
		log.Fatal(err)
	}

	//инициализируем хранилище
	var stor storage.Repository
	l.Info("Используем Postgre")
	//stor = storage.NewPostgreBase("urlshortnerdb", "postgres", "1101", "192.168.31.40", "5432", conf.Log.Logger)

	//ps := fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=disable", `192.168.31.40:5432`, `postgres`, `1101`, `urlshortnerdb`)

	//db, err := sql.Open("pgx", ps)
	//if err != nil {
	//	fmt.Println("Нет коннекта")
	//	panic(err)
	//}
	//defer db.Close()

	urlExample := "postgres://postgres:1101@192.168.31.40:5432/urlshortnerdb"
	db, err := pgx.Connect(context.Background(), urlExample)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close(context.Background())
	stor = storage.NewPostgreBase(db, l)

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
			l,
		},
	}

	l.Info("Running server", zap.String("address", conf.LocalAddress.String()), zap.String("loglevel", conf.LogLevel))

	if err := srv.Run(conf.LocalAddress); err != nil {
		log.Fatal(err)
	}

}
