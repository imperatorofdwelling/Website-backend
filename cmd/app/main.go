package main

import (
	"github.com/https-whoyan/dwellingPayload/config"
	"github.com/https-whoyan/dwellingPayload/internal/server/http"
	"log"
	"time"
)

func main() {
	// TODO run Server and DB from the config
	config.LoadConfig()
	go startServer()
	go initDB()
}

func startServer() {
	serverCfg := &config.Server{
		Addr:         "localhost:8080",
		ReadTimeout:  time.Second * 4,
		WriteTimeout: time.Second * 4,
		IdleTimeout:  time.Minute,
	}
	cfg := &config.Config{
		Server: serverCfg,
	}
	server := http.New(cfg.Server)
	log.Println("Server starting...")
	server.Run()
}

func initDB() {

}
