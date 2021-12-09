package main

import (
	"bufio"
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"os"

	pb "github.com/slntopp/nocloud-tunnel-mesh/pkg/proto"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func init() {
	viper.AutomaticEnv()
	viper.SetDefault("TUNNEL_HOST", "localhost:8080")
	viper.SetDefault("SECURE", true)
}

func main() {

	host := viper.GetString("TUNNEL_HOST")

	var opts []grpc.DialOption
	// opts = append(opts, grpc.WithInsecure())
	// if viper.GetBool("SECURE") {
	// 	cred := credentials.NewTLS(&tls.Config{InsecureSkipVerify: true})
	// 	opts[0] = grpc.WithTransportCredentials(cred)
	// }

	if viper.GetBool("SECURE") {
		// Load client cert
		 //cert, err := tls.LoadX509KeyPair("cert/0client.crt", "cert/0client.key")
		cert, err := tls.LoadX509KeyPair("cert/1client.crt", "cert/1client.key")
		if err != nil {
			log.Fatal("fail to LoadX509KeyPair:", err)
		}

		// // Load CA cert
		// //Certification authority, CA
		// //A CA certificate is a digital certificate issued by a certificate authority (CA), so SSL clients (such as web browsers) can use it to verify the SSL certificates sign by this CA.
		// caCert, err := ioutil.ReadFile("../cert/cacerts.cer")
		// if err != nil {
		// 	log.Fatal(err)
		// }
		// caCertPool := x509.NewCertPool()
		// caCertPool.AppendCertsFromPEM(caCert)

		// Setup HTTPS client
		config := &tls.Config{
			Certificates: []tls.Certificate{cert},
			// RootCAs:            caCertPool,
			// InsecureSkipVerify: false,
			InsecureSkipVerify: true,
		}
		// config.BuildNameToCertificate()
		cred := credentials.NewTLS(config)

		// cred := credentials.NewTLS(&tls.Config{InsecureSkipVerify: true})
		// opts[0] = grpc.WithTransportCredentials(cred)
		opts = append(opts, grpc.WithTransportCredentials(cred))
	} else {
	opts = append(opts, grpc.WithInsecure())
	}

	opts = append(opts, grpc.WithBlock())

	conn, err := grpc.Dial(host, opts...)
	if err != nil {
		log.Fatal("fail to dial:", err)
	}
	defer conn.Close()

	log.Println("Connected to server", host)

	client := pb.NewTunnelClient(conn)

	stdreader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("c2s > ")
		note, _ := stdreader.ReadString('\n')

		// if err := stream.Send(&pb.StreamData{Message: note}); err != nil {
		// 	lg.Fatal("Failed to send a note:", zap.Error(err))

		r, err := client.SendData(context.Background(), &pb.SendDataRequest{Host: "ClientZero", Message: note})
		if err != nil {
			log.Printf("could not SendData: %v", err)
		}
		log.Printf("Greeting c: %v", r.GetResult())

	}

}
