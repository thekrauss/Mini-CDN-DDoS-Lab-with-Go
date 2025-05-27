package main

import (
	"log"

	"github.com/thekrauss/Mini-CDN-DDoS-Lab-with-Go/control-plane/config"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failure to load configuration : %v", err)
	}
	config.AppConfig = *cfg

}
