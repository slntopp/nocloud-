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
	stream, err := client.InitTunnel(context.Background(), &pb.InitTunnelRequest{Host: "testo"})
	if err != nil {
		panic(err)
	}
	for {
		var msg string
		err := stream.RecvMsg(&msg)
		if err == io.EOF {
			return
		}
		if err != nil {
			panic(err)
		}
	}
}
