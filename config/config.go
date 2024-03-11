package config

import (
	"errors"
	"fmt"
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

//var ConfigGlobal = Config{
//	LocalAddress: LocalAddress{
//		IP:   "127.0.0.1",
//		Port: 8080,
//	},
//	BaseURL: "",
//}
