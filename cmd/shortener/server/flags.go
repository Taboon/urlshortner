package server

import (
	"flag"
	"fmt"
	"github.com/Taboon/urlshortner/cmd/shortener/config"
)

func parseFlags() {
	// регистрируем переменную flagRunAddr
	// как аргумент -a со значением :8080 по умолчанию
	flag.StringVar(&config.ConfigGlobal.BaseURL, "b", "localhost", "address to make short url")
	flag.Var(&config.ConfigGlobal.LocalAddress, "a", "address to start server")
	// парсим переданные серверу аргументы в зарегистрированные переменные
	flag.Parse()
	fmt.Println("Server started on: " + config.ConfigGlobal.URL())
}
