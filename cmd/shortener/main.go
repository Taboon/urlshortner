package main

import (
	"github.com/Taboon/urlshortner/cmd/shortener/server"
)

func main() {
	if err := server.Run(); err != nil {
		panic(err)
	}
}
