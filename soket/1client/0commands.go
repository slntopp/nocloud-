package client

import (
	"bufio"
	"log"
	"net"
	"strings"
)

//HandleClient помощник
func HandleClient(conn net.Conn) {

	// time.Sleep(5 * time.Second)

	conn.Write([]byte("Hello server!\n"))

	// buffer := make([]byte, 4096)
	// n, _ := conn.Read(buffer)
	// log.Print(string(buffer[:n]))
	// log.Println("Hello server!")

	buf := bufio.NewScanner(conn)
	for buf.Scan() {
		log.Println("From server:" + buf.Text())
		conn.Write([]byte("Hello from client! " + strings.ToUpper(buf.Text()) + "!\n"))
	}

}
