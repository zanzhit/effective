package config

import (
	"log"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

type Database struct {
	Host     string `env:"DB_HOST" env-default:"localhost"`
	User     string `env:"DB_USER" env-required:"true"`
	Password string `env:"DB_PASSWORD" env-required:"true"`
	Name     string `env:"DB_NAME" env-required:"true"`
	SSLMode  string `env:"DB_SSLMODE" env-required:"true"`
	Port     int    `env:"DB_PORT" env-required:"true"`
}

type HTTPServer struct {
	Address     string        `env:"ADDRESS" env-required:"true"`
	User        string        `env:"USER" env-required:"true"`
	Password    string        `env:"USER_PASSWORD" env-required:"true"`
	Timeout     time.Duration `env:"TIMEOUT" env-default:"4s"`
	IdleTimeout time.Duration `env:"IDLE_TIMEOUT" env-default:"60s"`
}

type Config struct {
	Env           string     `env:"ENV" env-default:"local"`
	ExternalAPI   string     `env:"EXTERNAL_API" env-required:"true"`
	PageSizeLimit int        `env:"PAGE_SIZE_LIMIT" env-default:"20"`
	DB            Database   `env:",embedded"`
	HTTPServer    HTTPServer `env:",embedded"`
}

func MustLoad() *Config {
	var cfg Config

	if err := godotenv.Load(); err != nil {
		log.Fatalf(".env file does not exist %s", err)
	}

	if err := cleanenv.ReadEnv(&cfg); err != nil {
		log.Fatalf("cannot read .env file config: %s", err)
	}

	return &cfg
}
