package config

import (
	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
	"log"
	"os"
	"time"
)

type Config struct {
	Server *Server `yaml:"server"`
	DB     *DB     `yaml:"db"`
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

type DB struct {
	DBHost     string `yaml:"db_host"`
	DBPort     string `yaml:"db_port"`
	DBUsername string `yaml:"db_username"`
	DBPassword string `yaml:"db_password"`
	DBName     string `yaml:"db_name"`
	DBSSLMode  string `yaml:"dbssl_mode"`
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
