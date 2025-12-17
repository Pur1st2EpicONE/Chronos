package config

import (
	"fmt"
	"os"
	"time"

	wbf "github.com/wb-go/wbf/config"
)

type Config struct {
	Logger   Logger   `mapstructure:"logger"`
	Notifier Notifier `mapstructure:"notifier"`
	Server   Server   `mapstructure:"server"`
	Storage  Storage  `mapstructure:"database"`
	Broker   Broker   `mapstructure:"broker"`
	Cache    Cache    `mapstructure:"cache"`
}

type Notifier struct {
	TelegramToken    string
	TelegramReceiver string
}

type Logger struct {
	Debug  bool   `mapstructure:"debug_mode"`
	LogDir string `mapstructure:"log_directory"`
}

type Server struct {
	Port            string        `mapstructure:"port"`
	ReadTimeout     time.Duration `mapstructure:"read_timeout"`
	WriteTimeout    time.Duration `mapstructure:"write_timeout"`
	MaxHeaderBytes  int           `mapstructure:"max_header_bytes"`
	ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout"`
}

type Broker struct {
	URL            string        `mapstructure:"url"`
	QueueName      string        `mapstructure:"queue_name"`
	ConnectionName string        `mapstructure:"connection_name"`
	ConnectTimeout time.Duration `mapstructure:"connect_timeout"`
	Heartbeat      time.Duration `mapstructure:"heartbeat"`
	Reconnect      Producer      `mapstructure:"reconnect"`
	Producer       Producer      `mapstructure:"producer"`
	Consumer       Consumer      `mapstructure:"consumer"`
}

type Producer struct {
	Attempts        int           `mapstructure:"attempts"`
	Delay           time.Duration `mapstructure:"delay"`
	Backoff         float64       `mapstructure:"backoff"`
	MessageQueueTTL time.Duration `mapstructure:"message_queue_ttl"`
}

type Consumer struct {
	Attempts      int           `mapstructure:"attempts"`
	Delay         time.Duration `mapstructure:"delay"`
	Backoff       float64       `mapstructure:"backoff"`
	ConsumerTag   string        `mapstructure:"consumer_tag"`
	AutoAck       bool          `mapstructure:"auto_ack"`
	Workers       int           `mapstructure:"workers"`
	PrefetchCount int           `mapstructure:"prefetch_count"`
}

type Storage struct {
	Host            string        `mapstructure:"host"`
	Port            string        `mapstructure:"port"`
	Username        string        `mapstructure:"username"`
	Password        string        `mapstructure:"password"`
	DBName          string        `mapstructure:"dbname"`
	SSLMode         string        `mapstructure:"sslmode"`
	MaxOpenConns    int           `mapstructure:"max_open_conns"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
}

type Cache struct {
	Host      string `mapstructure:"host"`
	Port      string `mapstructure:"port"`
	Password  string `mapstructure:"password"`
	MaxMemory string `mapstructure:"max_memory"`
	Policy    string `mapstructure:"policy"`
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

	conf.Notifier.TelegramToken = os.Getenv("TG_BOT_TOKEN")
	conf.Notifier.TelegramReceiver = os.Getenv("TG_CHAT_ID")

	conf.Storage.Username = os.Getenv("DB_USER")
	conf.Storage.Password = os.Getenv("DB_PASSWORD")

	conf.Cache.Password = os.Getenv("REDIS_PASSWORD")

	return conf, nil

}
