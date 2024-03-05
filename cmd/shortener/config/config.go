package config

import (
	"fmt"
	"strconv"
	"strings"
)

type Config struct {
	LocalAddress LocalAddress
	BaseUrl      string
}

type LocalAddress struct {
	Ip   string
	Port int
}

// String должен уметь сериализовать переменную типа в строку.
func (l *LocalAddress) String() string {
	var address = []string{
		l.Ip, fmt.Sprint(l.Port),
	}
	return fmt.Sprint(strings.Join(address, ":"))
}

// Set связывает переменную типа со значением флага
// и устанавливает правила парсинга для пользовательского типа.
func (l *LocalAddress) Set(flagValue string) error {
	address := strings.Split(flagValue, ":")
	l.Ip = address[0]
	port, err := strconv.Atoi(address[1])

	if err != nil {
		return err
	}
	l.Port = port
	return nil
}

func (c *Config) Url() string {
	return c.LocalAddress.Ip + ":" + strconv.Itoa(c.LocalAddress.Port)
}

var ConfigGlobal = Config{
	LocalAddress: LocalAddress{
		Ip:   "127.0.0.1",
		Port: 8080,
	},
	BaseUrl: "localhost",
}
