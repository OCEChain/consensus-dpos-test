package main

import (
	"github.com/joho/godotenv"
	"log"
	"os"
	"./blockchain"
)

func main() {
	err := godotenv.Load("dpos.env")
	if err != nil {
		log.Fatal(err)
	}
	command := os.Args[1:]
	nodeID := "3001"
	if len(command) >= 1 {
		nodeID = command[0]
		log.Println(nodeID)
	}
	blockchain.StartServer(nodeID)
}