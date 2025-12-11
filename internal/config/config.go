package config

import (
	"fmt"
	"os"
	"time"

	wbf "github.com/wb-go/wbf/config"
)

type Config struct {
	Logger  Logger  `mapstructure:"logger"`
	Server  Server  `mapstructure:"server"`
	Storage Storage `mapstructure:"database"`
	Broker  Broker  `mapstructure:"broker"`
}

type Logger struct {
	Level string `mapstructure:"level"`
}

type Server struct {
	Port            string        `mapstructure:"port"`
	ReadTimeout     time.Duration `mapstructure:"read_timeout"`
	WriteTimeout    time.Duration `mapstructure:"write_timeout"`
	MaxHeaderBytes  int           `mapstructure:"max_header_bytes"`
	ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout"`
}

type Broker struct {
	URL                  string        `mapstructure:"url"`
	ConnectionName       string        `mapstructure:"connection_name"`
	ConnectTimeout       time.Duration `mapstructure:"connect_timeout"`
	Heartbeat            time.Duration `mapstructure:"heartbeat"`
	MessageQueueTTL      time.Duration `mapstructure:"message_queue_ttl"`
	ConsumeRetryAttempts int           `mapstructure:"consume_retry_attempts"`
	ConsumeRetryDelay    time.Duration `mapstructure:"consume_retry_delay"`
	ConsumeRetryBackoff  float64       `mapstructure:"consume_retry_backoff"`
	PublishRetryAttempts int           `mapstructure:"publish_retry_attempts"`
	PublishRetryDelay    time.Duration `mapstructure:"publish_retry_delay"`
	PublishRetryBackoff  float64       `mapstructure:"publish_retry_backoff"`
}

type Storage struct {
	Host            string        `mapstructure:"host"`
	Port            string        `mapstructure:"port"`
	Username        string        `mapstructure:"username"`
	Password        string        `mapstructure:"passwod"`
	DBName          string        `mapstructure:"dbname"`
	SSLMode         string        `mapstructure:"sslmode"`
	MaxOpenConns    int           `mapstructure:"max_open_conns"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
}

func Load() (Config, error) {

	cfg := wbf.New()

	if err := cfg.LoadEnvFiles(".env"); err != nil {
		return Config{}, err
	}

	if err := cfg.LoadConfigFiles("./configs/config.yaml"); err != nil {
		return Config{}, err
	}

	var conf Config

	if err := cfg.Unmarshal(&conf); err != nil {
		return Config{}, fmt.Errorf("unmarshal config: %w", err)
	}

	conf.Storage.Username = os.Getenv("DB_USER")
	conf.Storage.Password = os.Getenv("DB_PASSWORD")

	return conf, nil

}
