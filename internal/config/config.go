// Package config provides the configuration structures and loading logic for the Chronos application.
// It supports loading configuration from YAML files, environment variables, and applies defaults.
package config

import (
	"fmt"
	"os"
	"time"

	wbf "github.com/wb-go/wbf/config"
)

// Config is the top-level application configuration, containing logger, notifier, server, storage, broker, and cache settings.
type Config struct {
	Logger   Logger   `mapstructure:"logger"`   // logger configuration
	Notifier Notifier `mapstructure:"notifier"` // notifier configuration
	Server   Server   `mapstructure:"server"`   // server configuration
	Storage  Storage  `mapstructure:"database"` // database/storage configuration
	Broker   Broker   `mapstructure:"broker"`   // broker configuration
	Cache    Cache    `mapstructure:"cache"`    // cache configuration
}

// Notifier contains credentials and settings for Telegram and Email notifications.
type Notifier struct {
	TelegramToken    string // Telegram bot token
	TelegramReceiver string // Telegram chat ID
	EmailSender      string // email sender address
	EmailPassword    string // email password
	EmailSMTP        string // SMTP server username/password if needed
	EmailSMTPAddr    string // SMTP server address
}

// Logger defines logging configuration.
type Logger struct {
	Debug  bool   `mapstructure:"debug_mode"`    // enable debug mode
	LogDir string `mapstructure:"log_directory"` // directory for log files
}

// Server contains HTTP server configuration.
type Server struct {
	Port            string        `mapstructure:"port"`             // server port
	ReadTimeout     time.Duration `mapstructure:"read_timeout"`     // HTTP read timeout
	WriteTimeout    time.Duration `mapstructure:"write_timeout"`    // HTTP write timeout
	MaxHeaderBytes  int           `mapstructure:"max_header_bytes"` // max header size
	ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout"` // graceful shutdown timeout
}

// Broker contains configuration for the message broker.
type Broker struct {
	URL                 string        `mapstructure:"url"`                  // broker connection URL
	QueueName           string        `mapstructure:"queue_name"`           // main queue name
	ConnectionName      string        `mapstructure:"connection_name"`      // connection name
	ConnectTimeout      time.Duration `mapstructure:"connect_timeout"`      // timeout for initial connection
	Reconnect           Producer      `mapstructure:"reconnect"`            // reconnect retry strategy
	Producer            Producer      `mapstructure:"producer"`             // producer retry strategy
	Consumer            Consumer      `mapstructure:"consumer"`             // consumer retry strategy
	CleanupInterval     time.Duration `mapstructure:"cleanup_interval"`     // interval for cleanup tasks
	HealthcheckInterval time.Duration `mapstructure:"healthcheck_interval"` // interval for health checks
}

// Producer defines retry and message queue settings for producer operations.
type Producer struct {
	Attempts        int           `mapstructure:"attempts"`          // number of retry attempts
	Delay           time.Duration `mapstructure:"delay"`             // delay between retries
	Backoff         float64       `mapstructure:"backoff"`           // backoff multiplier
	MessageQueueTTL time.Duration `mapstructure:"message_queue_ttl"` // queue TTL for messages
}

// Consumer defines retry and consumption settings for consumer operations.
type Consumer struct {
	Attempts      int           `mapstructure:"attempts"`       // number of retry attempts
	Delay         time.Duration `mapstructure:"delay"`          // delay between retries
	Backoff       float64       `mapstructure:"backoff"`        // backoff multiplier
	ConsumerTag   string        `mapstructure:"consumer_tag"`   // consumer tag
	AutoAck       bool          `mapstructure:"auto_ack"`       // auto-acknowledge messages
	Workers       int           `mapstructure:"workers"`        // number of worker goroutines
	PrefetchCount int           `mapstructure:"prefetch_count"` // prefetch count for QoS
}

// Storage defines database connection and query retry configuration.
type Storage struct {
	Host               string        `mapstructure:"host"`                 // DB host
	Port               string        `mapstructure:"port"`                 // DB port
	Username           string        `mapstructure:"username"`             // DB username
	Password           string        `mapstructure:"password"`             // DB password
	DBName             string        `mapstructure:"dbname"`               // database name
	SSLMode            string        `mapstructure:"sslmode"`              // SSL mode
	MaxOpenConns       int           `mapstructure:"max_open_conns"`       // maximum open connections
	MaxIdleConns       int           `mapstructure:"max_idle_conns"`       // maximum idle connections
	ConnMaxLifetime    time.Duration `mapstructure:"conn_max_lifetime"`    // max lifetime per connection
	RecoverLimit       int           `mapstructure:"recover_limit"`        // limit for recovery operations
	QueryRetryStrategy Producer      `mapstructure:"query_retry_strategy"` // query retry strategy
	RetentionStrategy  Retention     `mapstructure:"retention_strategy"`   // retention durations
}

// Retention specifies retention periods for notifications by status.
type Retention struct {
	Canceled  time.Duration // retention for canceled notifications
	Completed time.Duration // retention for completed notifications
	Failed    time.Duration // retention for failed notifications
}

// Cache defines Redis cache connection and retry configuration.
type Cache struct {
	Host           string        `mapstructure:"host"`            // cache host
	Port           string        `mapstructure:"port"`            // cache port
	Password       string        `mapstructure:"password"`        // cache password
	MaxMemory      string        `mapstructure:"max_memory"`      // max memory for Redis
	Policy         string        `mapstructure:"policy"`          // eviction policy
	RetryStrategy  Producer      `mapstructure:"retry_strategy"`  // retry strategy for cache operations
	ExpirationTime time.Duration `mapstructure:"expiration_time"` // key expiration duration
}

// Load reads configuration from YAML files and environment variables, applies defaults, and returns a Config instance.
func Load() (Config, error) {

	cfg := wbf.New()

	if err := cfg.LoadConfigFiles("./config.yaml"); err != nil {
		return Config{}, err
	}

	if err := cfg.LoadEnvFiles(".env"); err != nil && !cfg.GetBool("docker") {
		return Config{}, err
	}

	var conf Config

	if err := cfg.Unmarshal(&conf); err != nil {
		return Config{}, fmt.Errorf("unmarshal config: %w", err)
	}

	loadEnvs(&conf)

	return conf, nil

}

// loadEnvs overrides configuration fields with values from environment variables.
func loadEnvs(conf *Config) {

	conf.Storage.Username = os.Getenv("DB_USER")
	conf.Storage.Password = os.Getenv("DB_PASSWORD")

	conf.Cache.Password = os.Getenv("REDIS_PASSWORD")

	conf.Notifier.TelegramToken = os.Getenv("TG_BOT_TOKEN")
	conf.Notifier.TelegramReceiver = os.Getenv("TG_CHAT_ID")

	conf.Notifier.EmailSender = os.Getenv("GOOGLE_APP_EMAIL")
	conf.Notifier.EmailPassword = os.Getenv("GOOGLE_APP_PASSWORD")
	conf.Notifier.EmailSMTP = os.Getenv("GOOGLE_APP_SMTP")
	conf.Notifier.EmailSMTPAddr = os.Getenv("SMTP_ADDR")

}
