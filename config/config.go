package config

import (
	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
	"log"
	"os"
	"time"
)

type Config struct {
	Server Server `yaml:"server"`
}

type Server struct {
	Addr         string        `yaml:"addr"`
	ReadTimeout  time.Duration `yaml:"readTimeout"`
	WriteTimeout time.Duration `yaml:"writeTimeout"`
	/*
		Idle timeout is a period of time during which
		the server or connection waits for any action from the client.
	*/
	IdleTimeout time.Duration `yaml:"idleTimeout"`
}

func LoadConfig() *Config {
	// Loading .env vars
	if err := godotenv.Load("./env"); err != nil {
		log.Fatal(err)
	}

	cfg := new(Config)

	// Get path to cfg from .env
	cfgPath := os.Getenv("LOCAL_CFG_PATH")
	if err := cleanenv.ReadConfig(cfgPath, cfg); err != nil {
		log.Fatal(err)
	}
	return cfg
}
