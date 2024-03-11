package main

import (
	"github.com/Taboon/urlshortner/cmd/shortener/storage"
	"log"

	"github.com/Taboon/urlshortner/cmd/shortener/server"
)

func main() {
	stor := storage.TempStorage{}
	serv := server.Server{}
	conf, err := serv.ParseFlags()
	if err != nil {
		log.Fatal(err)
	}
	serv.Conf = conf
	serv.Stor = stor
	if err := serv.Run(); err != nil {
		log.Fatal(err)
	}
}
