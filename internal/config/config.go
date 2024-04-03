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
	LogLevel     string
	FileBase     FileBase
}

type Address struct {
	IP   string
	Port int
}

type FileBase struct {
	File string
}

func (f *FileBase) String() string {
	return f.File
}

func (f *FileBase) Set(flagValue string) error {
	if flagValue == "" {
		f.File = ""
		return nil
	}
	f.File = flagValue
	return nil
}

var emptyFlagError = errors.New("пустое значение флага")

const baseFilePath = "/tmp/short-url-db.json"

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
		return emptyFlagError
	}

	flagValue = strings.TrimPrefix(flagValue, "http://")
	flagValue = strings.TrimPrefix(flagValue, "https://")

	address := strings.Split(flagValue, ":")

	for i, v := range address {
		if i == 0 && v != "" {
			l.IP = v
		}
		if i == 1 && v != "" {
			port, err := strconv.Atoi(v)

			if err != nil {
				return err
			}
			l.Port = port
		}
	}

	return nil
}

func (c *Config) URL() string {
	return fmt.Sprintf("%v:%v", c.LocalAddress.IP, strconv.Itoa(c.LocalAddress.Port))
}

func BuildConfig() *Config {
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

	conf.FileBase.File = baseFilePath

	err := parseEnv(&conf)
	if err != nil {
		fmt.Println(err)
	}

	err = parseFlags(&conf)
	if err != nil {
		fmt.Println(err)
	}

	return &conf
}

func parseEnv(conf *Config) error {
	if envRunAddr := os.Getenv("SERVER_ADDRESS"); envRunAddr != "" {
		err := conf.LocalAddress.Set(envRunAddr)
		if err != nil {
			return err
		}
	}
	if envBasePath := os.Getenv("RUN_ADDR"); envBasePath != "" {
		err := conf.BaseURL.Set(envBasePath)
		if err != nil {
			return err
		}
	}
	if envBasePath := os.Getenv("FILE_STORAGE_PATH"); envBasePath != "" {
		conf.FileBase.File = envBasePath
	}
	return nil
}

func parseFlags(conf *Config) error {

	flag.Var(&conf.BaseURL, "b", "address to make short url")
	flag.Var(&conf.LocalAddress, "a", "address to start server")
	flag.Var(&conf.FileBase, "f", "file base path")
	flag.StringVar(&conf.LogLevel, "log", "Info", "loglevel (Info, Debug, Error)")

	flag.Parse()

	return nil
}
