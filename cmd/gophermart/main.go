package main

import (
	"github.com/ry461ch/loyalty_system/internal/config"
	"github.com/ry461ch/loyalty_system/internal/app"
)

func main() {
	server := server.NewServer(config.New())
	server.Run()
}