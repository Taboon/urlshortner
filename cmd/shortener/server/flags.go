package server

import (
	"flag"
	"fmt"
	"github.com/Taboon/urlshortner/cmd/shortener/config"
	"os"
)

func parseFlags() {
	// регистрируем переменную flagRunAddr
	// как аргумент -a со значением :8080 по умолчанию
	flag.StringVar(&config.ConfigGlobal.BaseURL, "b", "http://127.0.0.1:8080", "address to make short url")
	flag.Var(&config.ConfigGlobal.LocalAddress, "a", "address to start server")

	flag.Parse()
	if envRunAddr := os.Getenv("SERVER_ADDRESS"); envRunAddr != "" {
		err := config.ConfigGlobal.LocalAddress.Set(envRunAddr)
		if err != nil {
			fmt.Println(err)
		}
	}
	if envBasePath := os.Getenv("RUN_ADDR"); envBasePath != "" {
		config.ConfigGlobal.BaseURL = envBasePath
	}
	fmt.Println("Server started on: " + config.ConfigGlobal.URL())
	fmt.Println("Base URL: " + config.ConfigGlobal.BaseURL)
}
