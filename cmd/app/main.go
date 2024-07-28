package main

import (
	"github.com/imperatorofdwelling/Website-backend/config"
	_ "github.com/imperatorofdwelling/Website-backend/docs"
	"github.com/imperatorofdwelling/Website-backend/pkg/logger"
)

//	@title			Imperator Of Dwelling Payment System
//	@version		1.0
//	@description	Payment System for Imperator Of Dwelling written on Golang

//	@host		localhost:8080
//	@BasePath	/

func main() {
	cfg := config.LoadConfig("")
	log := logger.New(logger.EnvLocal)
	cfg.Run(log)
}
