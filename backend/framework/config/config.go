package config

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
	"reflect"
	"strconv"
)

type AppConfig struct {
	Host      string `yaml:"host" env:"HOST"`
	Port      string `yaml:"port" env:"PORT"`
	CacheType string `yaml:"cache_type" env:"CACHE_TYPE"`
	Db        struct {
		Host     string `yaml:"host" env:"DB_HOST"`
		Port     int    `yaml:"port" env:"DB_PORT"`
		Name     string `yaml:"name" env:"DB_NAME"`
		User     string `yaml:"user" env:"DB_USER"`
		Password string `yaml:"password" env:"DB_PASSWORD"`
	}
	JWT struct {
		Secret   string `yaml:"secret" env:"JWT_SECRET"`
		Lifetime int    `yaml:"lifetime" env:"JWT_LIFETIME"`
		Domain   string `yaml:"domain" env:"JWT_DOMAIN"`
	}
	Root struct {
		Mail     string `yaml:"mail" env:"ROOT_MAIL"`
		Password string `yaml:"password" env:"ROOT_PASSWORD"`
	}
	Cors struct {
		Origins string `yaml:"origins" env:"CORS_ORIGINS"`
	}
	Mail struct {
		Sender        string `yaml:"sender" env:"MAIL_SENDER"`
		Receiver      string `yaml:"receiver" env:"MAIL_RECEIVER"`
		MaxCount      int    `yaml:"max_count" env:"MAIL_MAX_COUNT"`
		CountDuration int    `yaml:"count_duration" env:"MAIL_COUNT_DURATION"`
		SMTP          struct {
			Host     string `yaml:"host" env:"MAIL_SMTP_HOST"`
			Port     int    `yaml:"port" env:"MAIL_SMTP_PORT"`
			User     string `yaml:"user" env:"MAIL_SMTP_USER"`
			Password string `yaml:"password" env:"MAIL_SMTP_PASSWORD"`
		}
		MaxMessageLength int `yaml:"max_message_length" env:"MAIL_MAX_MESSAGE_LENGTH"`
	}
	Location struct {
		Host                 string `yaml:"host" env:"LOCATION_HOST"`
		Account              string `yaml:"account" env:"LOCATION_ACCOUNT"`
		License              string `yaml:"license" env:"LOCATION_LICENSE"`
		CacheExpireHours     int    `yaml:"cache_expire_hours" env:"LOCATION_CACHE_EXPIRE_HOURS"`
		CheckWaitTimeMinutes int    `yaml:"check_wait_time_minutes" env:"LOCATION_CHECK_WAIT_TIME_MINUTES"`
		LocalIp              string `yaml:"local_ip" env:"LOCATION_LOCAL_IP"`
	}
	APM struct {
		ServiceName string `yaml:"service_name" env:"ELASTIC_APM_SERVICE_NAME"`
		SecretToken string `yaml:"secret_token" env:"ELASTIC_APM_SECRET_TOKEN"`
		ServerUrl   string `yaml:"server_url" env:"ELASTIC_APM_SERVER_URL"`
		ApiKey      string `yaml:"api_key" env:"ELASTIC_APM_API_KEY"`
		Environment string `yaml:"environment" env:"ELASTIC_APM_ENVIRONMENT"`
	}
	Log struct {
		Level    int    `yaml:"level" env:"LOG_LEVEL"`
		FileName string `yaml:"file_name" env:"LOG_FILE_NAME"`
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

	overwriteConfigWithEnvVars(&cachedConfig)
	return &cachedConfig
}

func overwriteConfigWithEnvVars(config *AppConfig) {
	reflectedConfig := reflect.ValueOf(config)
	if reflectedConfig.Kind() != reflect.Ptr || reflectedConfig.IsNil() {
		return
	}

	element := reflectedConfig.Elem()
	if element.Kind() != reflect.Struct {
		return
	}

	processStruct(element)
}

func processStruct(value reflect.Value) {
	valueType := value.Type()

	for i := 0; i < value.NumField(); i++ {
		field := value.Field(i)
		fieldType := valueType.Field(i)

		if field.Kind() == reflect.Struct {
			processStruct(field)
			continue
		}

		if fieldType.Tag.Get("env") != "" {
			envVar := fieldType.Tag.Get("env")
			envValue := os.Getenv(envVar)

			if envValue != "" {
				switch field.Kind() {
				case reflect.String:
					field.SetString(envValue)
				case reflect.Bool:
					field.SetBool(envValue == "true" || envValue == "1")
				case reflect.Int:
					num, err := strconv.Atoi(envValue)
					if err != nil {
						panic(fmt.Sprintf("could not parse env var '%s' as int, %v", envVar, err))
					}
					field.SetInt(int64(num))
				default:
					panic(fmt.Sprintf("unsupported field type '%s' for env var '%s'", field.Kind(), envVar))
				}
			}
		}
	}
}
