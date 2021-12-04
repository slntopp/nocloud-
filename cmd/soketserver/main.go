package main

import (
	"log"
	"net"

	server "github.com/EwRvp7LV7/nocloud-tunnel-mesh/soket/0server"
)

const (
	PORT = "10000"
)

func run() (err error) {

	lstnr, err := net.Listen("tcp", ":"+PORT)
	if err != nil {
		return err
	}

	log.Println("TCP server is UP @ localhost: " + PORT)
	defer lstnr.Close()

	for {
		//TODO Add limit queue/dispatcher
		connection, err := lstnr.Accept()
		if err != nil {
			log.Println("Client Connection failed")
			continue
		}

		go server.HandleServer(connection)
	}

	// return nil
}

func main() {

	if err := run(); err != nil {
		log.Fatal(err)
	}

}
