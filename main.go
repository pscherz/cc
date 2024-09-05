package main

import (
	"flag"
	"log"

	"github.com/pscherz/cc/client"
	"github.com/pscherz/cc/server"
)

var (
	flagServer  bool
	flagAddress string
)

func init() {
	flag.BoolVar(&flagServer, "server", false, "run server")
	flag.StringVar(&flagAddress, "address", ":4324", "if server, interface and port to serve from. if client, address to connect to server.")
	flag.Parse()
}

func main() {
	if flagServer {
		run_server()
	} else {
		run_client()
	}
}

func run_server() {
	log.Print("Running as server...")
	server.Init(flagAddress)
}

func run_client() {
	log.Print("Running as client...")
	client.Init(flagAddress)
}
