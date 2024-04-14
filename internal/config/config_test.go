package config

import (
	"os"
	"testing"
)

const baseFilePath = "/tmp/short-url-db.json"

func TestBuildConfig(t *testing.T) {
	// Mock environment variables
	err := os.Setenv("server_address", "127.0.0.1:8081")
	if err != nil {
		return
	}
	err = os.Setenv("run_addr", "http://localhost:8081")
	if err != nil {
		return
	}

	// Mock command line flags
	os.Args = []string{"", "-b", "http://localhost:8081", "-a", "127.0.0.1:8081", "-log", "debug"}

	configBuilder := NewConfigBuilder()
	configBuilder.SetLocalAddress("127.0.0.1", 8080)
	configBuilder.SetBaseURL("127.0.0.1", 8080)
	configBuilder.SetFileBase(baseFilePath)
	configBuilder.SetLogger("Info")
	configBuilder.ParseEnv()
	configBuilder.ParseFlag()
	conf := configBuilder.Build()

	// Verify configuration values
	if conf.LocalAddress.String() != "127.0.0.1:8081" {
		t.Errorf("Expected local address to be 127.0.0.1:8081, got %s", conf.LocalAddress.String())
	}

	if conf.BaseURL.String() != "localhost:8081" {
		t.Errorf("Expected base URL to be localhost:8081, got %s", conf.BaseURL.String())
	}

	if conf.LogLevel != "debug" {
		t.Errorf("Expected log level to be debug, got %s", conf.LogLevel)
	}
}
