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
	DataDase     string
	LogLevel     string
}

type ConfigBuilder interface {
	SetLocalAddress(ip string, port int) ConfigBuilder
	SetBaseURL(ip string, port int) ConfigBuilder
	SetFileBase(path string) ConfigBuilder
	SetLogger(level string) ConfigBuilder
	ParseFlag() ConfigBuilder
	ParseEnv() ConfigBuilder
	Build() *Config
}

type configBuilder struct {
	config *Config
}

func (c *configBuilder) SetLocalAddress(ip string, port int) ConfigBuilder {
	c.config.LocalAddress.IP = ip
	c.config.LocalAddress.Port = port
	return c
}

func (c *configBuilder) SetBaseURL(ip string, port int) ConfigBuilder {
	c.config.BaseURL.IP = ip
	c.config.BaseURL.Port = port
	return c
}

func (c *configBuilder) SetFileBase(path string) ConfigBuilder {
	c.config.FileBase.File = path
	return c
}

func (c *configBuilder) SetLogger(level string) ConfigBuilder {
	c.config.LogLevel = level
	return c
}

func (c *configBuilder) ParseFlag() ConfigBuilder {
	err := parseEnv(c.config)
	if err != nil {
		fmt.Println(err)
	}
	return c
}

func (c *configBuilder) ParseEnv() ConfigBuilder {
	err := parseFlags(c.config)
	if err != nil {
		fmt.Println(err)
	}
	return c
}

func (c *configBuilder) Build() *Config {
	return c.config
}
func NewConfigBuilder() ConfigBuilder {
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
		conf.DataDase = envDBAddres
	}
	return nil
}

func parseFlags(conf *Config) error {
	flag.Var(&conf.BaseURL, "b", "address to make short url")
	flag.StringVar(&conf.DataDase, "d", "", "data base url")
	flag.Var(&conf.LocalAddress, "a", "address to start server")
	flag.Var(&conf.FileBase, "f", "file base path")
	flag.StringVar(&conf.LogLevel, "log", "Info", "loglevel (Info, Debug, Error)")
	flag.Parse()
	return nil
}
