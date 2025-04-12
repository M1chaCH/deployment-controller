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
	Cors struct {
		Origins string `yaml:"origins"`
	}
	Mail struct {
		Sender        string `yaml:"sender"`
		Receiver      string `yaml:"receiver"`
		MaxCount      int    `yaml:"max_count"`
		CountDuration int    `yaml:"count_duration"`
		SMTP          struct {
			Host     string `yaml:"host"`
			Port     int    `yaml:"port"`
			User     string `yaml:"user"`
			Password string `yaml:"password"`
		}
		MaxMessageLength int `yaml:"max_message_length"`
	}
	Location struct {
		Host                 string `yaml:"host"`
		Account              string `yaml:"account"`
		License              string `yaml:"license"`
		CacheExpireHours     int    `yaml:"cache_expire_hours"`
		CheckWaitTimeMinutes int    `yaml:"check_wait_time_minutes"`
		LocalIp              string `yaml:"local_ip"`
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
