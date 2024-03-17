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
func (l Address) String() string {
	var address = []string{
		l.IP, fmt.Sprint(l.Port),
	}
	return fmt.Sprint(strings.Join(address, ":"))
}

// Set связывает переменную типа со значением флага
// и устанавливает правила парсинга для пользовательского типа.
func (l Address) Set(flagValue string) error {

	flagValue = strings.TrimPrefix(flagValue, "http://")
	flagValue = strings.TrimPrefix(flagValue, "https://")

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

func (c Config) URL() string {
	return fmt.Sprintf("%v:%v", c.LocalAddress.IP, strconv.Itoa(c.LocalAddress.Port))
}

func (c Config) BuildConfig() Config {
	conf := Config{
		//LocalAddress: Address{
		//	"127.0.0.1",
		//	8080,
		//},
		//BaseURL: Address{
		//	"127.0.0.1",
		//	8080,
		//},
	}

	err := c.parseEnv(conf)
	if err != nil {
		fmt.Println(err)
	}

	err = c.parseFlags(conf)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("Config:", conf)

	return conf
}

func (c Config) parseEnv(conf Config) error {
	if envRunAddr := os.Getenv("SERVER_ADDRESS"); envRunAddr != "" {
		err := conf.LocalAddress.Set(envRunAddr)
		if err != nil {
			return err
		}
		if envBasePath := os.Getenv("RUN_ADDR"); envBasePath != "" {
			err := conf.BaseURL.Set(envBasePath)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (c Config) parseFlags(conf Config) error {

	flag.Var(&conf.BaseURL, "b", "address to make short url")
	flag.Var(&conf.LocalAddress, "a", "address to start server")

	flag.Parse()

	fmt.Printf("Server started on: %v:%v\n", conf.LocalAddress.IP, conf.LocalAddress.Port)
	fmt.Printf("Base URL: %v:%v\n", conf.BaseURL.IP, conf.BaseURL.Port)

	return nil
}
