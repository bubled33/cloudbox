package main

import (
	"log"

	"github.com/yourusername/cloud-file-storage/internal/api"
)

func main() {

	server := api.NewServer()

	err := server.Run(":8080")
	if err != nil {
		log.Panicf("Server has error!")
	}
}
