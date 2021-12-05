package main

import (
	"context"
	"io"
	"log"

	pb "github.com/slntopp/nocloud-tunnel-mesh/pkg/proto"
	"google.golang.org/grpc"
)

func main() {
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())
	opts = append(opts, grpc.WithBlock())

	conn, err := grpc.Dial("localhost:10000", opts...)
	if err != nil {
		log.Fatalf("fail to dial: %v", err)
	}
	defer conn.Close()

	client := pb.NewTunnelClient(conn)

	stream, err := client.InitTunnel(context.Background()) //, &pb.InitTunnelRequest{Host: "testo"})
	if err != nil {
		log.Fatalf("%v.RouteChat(_) = _, %v", client, err)
	}
	// if err != nil {
	// 	panic(err)
	// }
	// for {
	// 	var msg string
	// 	err := stream.RecvMsg(&msg)
	// 	if err == io.EOF {
	// 		return
	// 	}
	// 	if err != nil {
	// 		panic(err)
	// 	}
	// }

	if err := stream.Send(&pb.InitTunnelRequest{Host: "Hello, server!"}); err != nil {
		log.Fatalf("Failed to send a note: %v", err)
	}

	for {
		in, err := stream.Recv()
		if err == io.EOF {
			return
		}
		if err != nil {
			log.Fatalf("Failed to receive a note : %v", err)
		}

		log.Printf("Hello from server! %s", in.Message)

		if err := stream.Send(&pb.InitTunnelRequest{Host: in.Message}); err != nil {
			log.Fatalf("Failed to send a note: %v", err)
		}

	}

}
