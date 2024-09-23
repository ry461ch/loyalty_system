package main

import (
	"github.com/ry461ch/loyalty_system/internal/app"
	"github.com/ry461ch/loyalty_system/internal/config"
)

func main() {
	server := server.NewServer(config.New())
	server.Run()
}
