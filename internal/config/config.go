package config

import (
	"fmt"
	"time"

	wbf "github.com/wb-go/wbf/config"
)

type Root struct {
	App App `mapstructure:"app"`
}

type App struct {
	Server  Server  `mapstructure:"server"`
	Service Service `mapstructure:"service"`
	Storage Storage `mapstructure:"storage"`
}

type Server struct {
	Port            string        `mapstructure:"port"`
	ReadTimeout     time.Duration `mapstructure:"read_timeout"`
	WriteTimeout    time.Duration `mapstructure:"write_timeout"`
	MaxHeaderBytes  int           `mapstructure:"max_header_bytes"`
	ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout"`
}

type Service struct {
}

type Storage struct {
	MasterDSN       string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

func Load() (App, error) {

	cfg := wbf.New()

	// if err := cfg.LoadEnvFiles(".env"); err != nil {
	// 	return App{}, err
	// }

	if err := cfg.LoadConfigFiles("./configs/config.yaml"); err != nil {
		fmt.Println("ASD")
		return App{}, err
	}

	cfg.EnableEnv("")

	var root Root
	if err := cfg.Unmarshal(&root); err != nil {
		return App{}, fmt.Errorf("unmarshal config: %w", err)
	}

	return root.App, nil

}
