package main

import (
	"github.com/Taboon/urlshortner/config"
	"github.com/Taboon/urlshortner/internal/server"
	"github.com/Taboon/urlshortner/internal/storage"
	"log"
)

func main() {
	stor := storage.TempStorage{}
	serv := server.Server{}
	conf := config.Config{}
	conf = conf.BuildConfig()
	serv.Conf = conf
	serv.Stor = stor
	if err := serv.Run(); err != nil {
		log.Fatal(err)
	}
}
