package main

import (
	"fmt"

	"github.com/harshjoeyit/goredis/server"
)

// go run main.go starts server
func main() {
	fmt.Println("Started go-redis")
	server.RunSyncTCPServer()
}
