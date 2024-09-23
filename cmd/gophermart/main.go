package main

import (
	"fmt"

	"github.com/ry461ch/loyalty_system/internal/app"
	"github.com/ry461ch/loyalty_system/internal/config"
)

func main() {
	fmt.Println("=================================================================")
	server := server.NewServer(config.New())
	server.Run()
}
