package config

import (
	"flag"
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	LocalAddress Address
	BaseURL      Address
	FileBase     FileBase
	DataBase     string
	LogLevel     string
	SecretKey    string
}

type Builder interface {
	SetLocalAddress(ip string, port int) Builder
	SetBaseURL(ip string, port int) Builder
	SetFileBase(path string) Builder
	SetLogger(level string) Builder
	ParseFlag() Builder
	ParseEnv() Builder
	Build() *Config
}

type configBuilder struct {
	config *Config
}

func (c *configBuilder) SetLocalAddress(ip string, port int) Builder {
	c.config.LocalAddress.IP = ip
	c.config.LocalAddress.Port = port
	return c
}

func (c *configBuilder) SetBaseURL(ip string, port int) Builder {
	c.config.BaseURL.IP = ip
	c.config.BaseURL.Port = port
	return c
}

func (c *configBuilder) SetFileBase(path string) Builder {
	c.config.FileBase.File = path
	return c
}

func (c *configBuilder) SetLogger(level string) Builder {
	c.config.LogLevel = level
	return c
}

func (c *configBuilder) ParseFlag() Builder {
	err := parseEnv(c.config)
	if err != nil {
		fmt.Println(err)
	}
	return c
}

func (c *configBuilder) ParseEnv() Builder {
	err := parseFlags(c.config)
	if err != nil {
		fmt.Println(err)
	}
	return c
}

func (c *configBuilder) Build() *Config {
	return c.config
}
func NewConfigBuilder() Builder {
	return &configBuilder{
		config: &Config{},
	}
}

func (c *Config) URL() string {
	return fmt.Sprintf("%v:%v", c.LocalAddress.IP, strconv.Itoa(c.LocalAddress.Port))
}

func parseEnv(conf *Config) error {
	if envRunAddr := os.Getenv("SERVER_ADDRESS"); envRunAddr != "" {
		err := conf.LocalAddress.Set(envRunAddr)
		if err != nil {
			return err
		}
	}
	if envBasePath := os.Getenv("BASE_URL"); envBasePath != "" {
		err := conf.BaseURL.Set(envBasePath)
		if err != nil {
			return err
		}
	}
	if envBasePath := os.Getenv("FILE_STORAGE_PATH"); envBasePath != "" {
		conf.FileBase.File = envBasePath
	}
	if envDBAddres := os.Getenv("DATABASE_DSN"); envDBAddres != "" {
		conf.DataBase = envDBAddres
	}
	if secretKey := os.Getenv("SECRET_KEY"); secretKey != "" {
		conf.SecretKey = secretKey
	}
	if fileBase := os.Getenv("TMP_FILE_BASE"); fileBase != "" {
		conf.SecretKey = fileBase
	}
	return nil
}

func parseFlags(conf *Config) error {
	flag.Var(&conf.BaseURL, "b", "address to make short url")
	flag.StringVar(&conf.DataBase, "d", "", "data base url")
	flag.Var(&conf.LocalAddress, "a", "address to start server")
	flag.Var(&conf.FileBase, "f", "file base path")
	flag.StringVar(&conf.LogLevel, "log", "Debug", "loglevel (Info, Debug, Error)")
	flag.Parse()
	return nil
}

func SetConfig() *Config {
	configBuilder := NewConfigBuilder()
	configBuilder.SetLocalAddress("127.0.0.1", 8080)
	configBuilder.SetBaseURL("127.0.0.1", 8080)
	configBuilder.SetLogger("Debug")
	configBuilder.ParseEnv()
	configBuilder.ParseFlag()
	conf := configBuilder.Build()
	return conf
}
