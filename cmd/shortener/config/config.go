package config

import (
	"fmt"
	"strconv"
	"strings"
)

type Config struct {
	LocalAddress LocalAddress
	BaseURL      string
}

type LocalAddress struct {
	IP   string
	Port int
}

// String должен уметь сериализовать переменную типа в строку.
func (l *LocalAddress) String() string {
	var address = []string{
		l.IP, fmt.Sprint(l.Port),
	}
	return fmt.Sprint(strings.Join(address, ":"))
}

// Set связывает переменную типа со значением флага
// и устанавливает правила парсинга для пользовательского типа.
func (l *LocalAddress) Set(flagValue string) error {
	address := strings.Split(flagValue, ":")
	l.IP = address[0]
	port, err := strconv.Atoi(address[1])

	if err != nil {
		return err
	}
	l.Port = port
	return nil
}

func (c *Config) Url() string {
	return c.LocalAddress.IP + ":" + strconv.Itoa(c.LocalAddress.Port)
}

var ConfigGlobal = Config{
	LocalAddress: LocalAddress{
		IP:   "127.0.0.1",
		Port: 8080,
	},
	BaseURL: "localhost",
}
