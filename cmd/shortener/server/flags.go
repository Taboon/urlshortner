package server

import (
	"flag"
	"fmt"
	"github.com/Taboon/urlshortner/cmd/shortener/config"
)

var conf = config.ConfigGlobal
var localAddress = conf.LocalAddress

func parseFlags() {
	// регистрируем переменную flagRunAddr
	// как аргумент -a со значением :8080 по умолчанию
	flag.StringVar(&conf.BaseUrl, "b", "localhost", "address to make short url")
	flag.Var(&localAddress, "a", "address to start server")
	// парсим переданные серверу аргументы в зарегистрированные переменные
	flag.Parse()
	fmt.Println("Server started on: " + conf.Url())
}
