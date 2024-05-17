package config

import (
	"log"

	"github.com/joho/godotenv"

	"github.com/https-whoyan/dwellingPayload/internal/server/http"
	"github.com/https-whoyan/dwellingPayload/pkg/repository"
)

type Config struct {
	Server            *http.ServerConfig            `yaml:"server"`
	PostgresSQLConfig *repository.PostgresSQLConfig `yaml:"db"`
}

func LoadConfig() *Config {
	err := loadDotEnv()
	if err != nil {
		//log Fatal by logger
		log.Fatal(err)
	}
	serverConfig, err := http.LoadConfig()
	if err != nil {
		//log Fatal by logger
		log.Fatal(err)
	}
	postgresSQLConfig, err := repository.LoadConfig()
	if err != nil {
		//log Fatal by logger
		log.Fatal(err)
	}
	cfg := &Config{
		Server:            serverConfig,
		PostgresSQLConfig: postgresSQLConfig,
	}
	return cfg
}

func (c *Config) Run() {
	err := repository.InitPostgresDB(c.PostgresSQLConfig)
	if err != nil {
		//log Fatal by logger
		log.Fatal(err)
	}
	srv := http.New(c.Server)

	defer c.Disconnect(srv)
	srv.Run()
}

func (c *Config) Disconnect(server *http.Server) {
	err := repository.Disconnect()
	if err != nil {
		// Log print by logger
		log.Println(err)
	}
	err = server.Disconnect()
	if err != nil {
		// Log print by logger
		log.Println(err)
	}
}

func loadDotEnv() error {
	err := godotenv.Load()
	return err
}
