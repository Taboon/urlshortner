package main

import (
	"github.com/Taboon/urlshortner/internal/config"
	"github.com/Taboon/urlshortner/internal/server"
	"github.com/Taboon/urlshortner/internal/storage"
	"log"
)

func main() {
	stor := storage.TempStorage{}
	srv := server.Server{}
	conf := config.Config{}
	conf = conf.BuildConfig()
	srv.Conf = conf
	srv.Stor = stor
	if err := srv.Run(); err != nil {
		log.Fatal(err)
	}
}
