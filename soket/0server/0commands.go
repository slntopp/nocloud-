package server

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"time"
)

func HandleServer(conn net.Conn) {
	defer conn.Close()

	go func() {
		stdreader := bufio.NewReader(os.Stdin)
		for {
			uname, _ := stdreader.ReadString('\n')
			_, err := conn.Write([]byte(uname))
			if err != nil {
				log.Println("Bye, client!2")
				return
			}
		}
	}()

	buf := bufio.NewScanner(conn)
	for buf.Scan() {
		conn.SetDeadline(time.Now().Add(time.Minute * 10))
		log.Println(buf.Text())
		fmt.Printf("m2c > ")
	}
	log.Println("Bye, client!")

}

// c := make(chan string)

// go func(c chan string) {
// 	buf := bufio.NewScanner(conn)
// 	for buf.Scan() {
// 		conn.SetDeadline(time.Now().Add(time.Minute * 10))
// 		log.Println(buf.Text())
// 		c <- "run"
// 	}
// 	c <- "stop"
// }(c)

// // go func() {
// // 	reader := bufio.NewReader(conn)
// // 	for {
// // 		conn.SetDeadline(time.Now().Add(time.Minute * 10))

// // 		line, err := reader.ReadString('\n')
// // 		if err != nil {
// // 			if err != io.EOF {
// // 				log.Println(err)
// // 			} else {
// // 				conn.Close()
// // 			}
// // 			break
// // 		}

// // 		log.Print(line)
// // 		fmt.Print("m2c > ")
// // 	}

// // }()

// stdreader := bufio.NewReader(os.Stdin)
// for {
// 	if "stop" == <-c {
// 		log.Println("Bye, client!1")
// 		return
// 	}

// 	fmt.Printf("m2c > ")
// 	uname, _ := stdreader.ReadString('\n')
// 	_, err := conn.Write([]byte(uname))
// 	if err != nil {
// 		log.Println("Bye, client!2")
// 		return
// 	}
// }
