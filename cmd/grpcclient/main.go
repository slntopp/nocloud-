package main

import (
	"context"
	"flag"
	"io"
	"log"

	pb "github.com/slntopp/nocloud-tunnel-mesh/grpcstream"
	"google.golang.org/grpc"
)

var (
	serverAddr         = flag.String("server_addr", "localhost:10000", "The server address in the format of host:port")
	serverHostOverride = flag.String("server_host_override", "x.test.example.com", "The server name used to verify the hostname returned by the TLS handshake")
)

// runRouteChat receives a sequence of route notes, while sending notes for various locations.
func runRouteChat1(client pb.RouteGuideClient) {
	// ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	// defer cancel()
	// stream, err := client.RouteChat(ctx)
	stream, err := client.RouteChat(context.Background())
	// stream, err := client.RouteChat() not enough arguments in call to client.RouteChat

	if err != nil {
		log.Fatalf("%v.RouteChat(_) = _, %v", client, err)
	}
	// waitc := make(chan struct{})

	// go func() {
	for {
		in, err := stream.Recv()
		if err == io.EOF {
			// read done.
			// close(waitc)
			return
		}
		if err != nil {
			log.Fatalf("Failed to receive a note : %v", err)
		}

		log.Printf("Hello from server! %s", in.Message)
		// log.Printf("Got message %s at point(%d, %d)", in.Message, in.Location.Latitude, in.Location.Longitude)

		notes := []*pb.RouteNote{
			{Message: in.Message},
		}
		for _, note := range notes {
			if err := stream.Send(note); err != nil {
				log.Fatalf("Failed to send a note: %v", err)
			}
		}

	}
	// }()

	// 	notes := []*pb.RouteNote{
	// 	// {Location: &pb.Point{Latitude: 0, Longitude: 1}, Message: in.Message},
	// 	{Location: &pb.Point{Latitude: 0, Longitude: 2}, Message: "Second message"},
	// 	{Location: &pb.Point{Latitude: 0, Longitude: 3}, Message: "Third message"},
	// 	{Location: &pb.Point{Latitude: 0, Longitude: 1}, Message: "Fourth message"},
	// 	{Location: &pb.Point{Latitude: 0, Longitude: 2}, Message: "Fifth message"},
	// 	{Location: &pb.Point{Latitude: 0, Longitude: 3}, Message: "Sixth message"},
	// }
	// for _, note := range notes {
	// 	if err := stream.Send(note); err != nil {
	// 		log.Fatalf("Failed to send a note: %v", err)
	// 	}
	// }

	// stream.CloseSend()
	// <-waitc
}

func main() {
	flag.Parse()
	var opts []grpc.DialOption

	opts = append(opts, grpc.WithInsecure())

	opts = append(opts, grpc.WithBlock())
	conn, err := grpc.Dial(*serverAddr, opts...)
	if err != nil {
		log.Fatalf("fail to dial: %v", err)
	}
	defer conn.Close()
	client := pb.NewRouteGuideClient(conn)

	// RouteChat
	runRouteChat1(client)
}
