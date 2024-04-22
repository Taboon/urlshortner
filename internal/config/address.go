package config

import (
	"fmt"
	"github.com/Taboon/urlshortner/internal/entity"
	"strconv"
	"strings"
)

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
	if flagValue == "" {
		return entity.ErrEmptyFlag
	}

	flagValue = strings.TrimPrefix(flagValue, "http://")
	flagValue = strings.TrimPrefix(flagValue, "https://")

	address := strings.Split(flagValue, ":")

	if len(address) > 1 {
		l.IP = address[0]
		port, err := strconv.Atoi(address[1])
		if err != nil {
			return err
		}
		l.Port = port
	}

	return nil
}
