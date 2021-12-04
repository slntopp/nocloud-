package main

import (
	"fmt"
	"net"

	client "github.com/EwRvp7LV7/nocloud-tunnel-mesh/soket/1client"
)

const (
	PORT = "10000"
	HOST = "localhost"
)

func run() (err error) {

	connect, err := net.Dial("tcp", HOST+":"+PORT)
	if err != nil {
		return err
	}
	defer connect.Close()

	fmt.Println("TCP server is Connected @ ", HOST, ":", PORT)

	client.HandleClient(connect)

	return nil
}

func main() {

	// flag.Parse()

	if err := run(); err != nil {
		fmt.Print(err)
	}

}
