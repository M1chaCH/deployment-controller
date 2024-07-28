package framework

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
)

type AppConfig struct {
	Host      string `yaml:"host"`
	Port      string `yaml:"port"`
	CacheType string `yaml:"cache_type"`
	Db        struct {
		Host     string `yaml:"host"`
		Port     int    `yaml:"port"`
		Name     string `yaml:"name"`
		User     string `yaml:"user"`
		Password string `yaml:"password"`
	}
	JWT struct {
		Secret   string `yaml:"secret"`
		Lifetime int    `yaml:"lifetime"`
		Domain   string `yaml:"domain"`
	}
	Root struct {
		Mail     string `yaml:"mail"`
		Password string `yaml:"password"`
	}
}

var cachedConfig AppConfig

func Config() *AppConfig {
	if (AppConfig{}) != cachedConfig {
		return &cachedConfig
	}

	data, err := os.ReadFile("config.yml")
	if err != nil {
		panic(fmt.Sprintf("config file could not be read, %v", err))
	}

	err = yaml.Unmarshal(data, &cachedConfig)
	if err != nil {
		panic(fmt.Sprintf("config file could not be parsed, %v", err))
	}

	return &cachedConfig
}
