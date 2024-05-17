package main

import (
	"github.com/https-whoyan/dwellingPayload/config"
)

func main() {
	cfg := config.LoadConfig()
	cfg.Run()
}
