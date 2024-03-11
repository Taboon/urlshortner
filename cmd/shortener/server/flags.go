package server

import (
	"flag"
	"fmt"
	"github.com/Taboon/urlshortner/config"
	"os"
)

func (s *Server) ParseFlags() (config.Config, error) {
	conf := config.Config{}

	if envRunAddr := os.Getenv("SERVER_ADDRESS"); envRunAddr != "" {
		err := conf.LocalAddress.Set(envRunAddr)
		if err != nil {
			fmt.Println(err)
		}
	}

	if envBasePath := os.Getenv("RUN_ADDR"); envBasePath != "" {
		err := conf.BaseURL.Set(envBasePath)
		if err != nil {
			return conf, err
		}
	}

	flag.Var(&conf.BaseURL, "b", "address to make short url")
	flag.Var(&conf.LocalAddress, "a", "address to start server")

	flag.Parse()

	fmt.Println("Server started on: " + conf.URL())
	fmt.Println("Base URL: ", conf.BaseURL)

	return conf, nil
}
