package config

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	LocalAddress Address
	BaseURL      Address
}

type Address struct {
	IP   string
	Port int
}

// String должен уметь сериализовать переменную типа в строку.
func (l *Address) String() string {
	var address = []string{
		l.IP, fmt.Sprint(l.Port),
	}
	return fmt.Sprint(strings.Join(address, ":"))
}

// Set связывает переменную типа со значением флага
// и устанавливает правила парсинга для пользовательского типа.
func (l *Address) Set(flagValue string) error {
	address := strings.Split(flagValue, ":")
	if address[0] == "" {
		err := errors.New("wrong adress")
		return err
	}
	l.IP = address[0]

	if address[1] == "" {
		err := errors.New("wrong port")
		return err
	}
	port, err := strconv.Atoi(address[1])

	if err != nil {
		return err
	}
	l.Port = port
	return nil
}

func (c *Config) URL() string {
	return c.LocalAddress.IP + ":" + strconv.Itoa(c.LocalAddress.Port)
}

func ParseFlags() (Config, error) {
	conf := Config{
		LocalAddress: Address{
			"127.0.0.1",
			8080,
		},
		BaseURL: Address{
			"127.0.0.1",
			8080,
		},
	}

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

	fmt.Printf("Server started on: %v:%v\n", conf.LocalAddress.IP, conf.LocalAddress.Port)
	fmt.Printf("Base URL: %v:%v\n", conf.BaseURL.IP, conf.BaseURL.Port)

	return conf, nil
}
